package agent

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/fantasy"
	"github.com/swadhinbiswas/ghost/internal/agent/hyper"
	"github.com/swadhinbiswas/ghost/internal/agent/notify"
	"github.com/swadhinbiswas/ghost/internal/agent/prompt"
	"github.com/swadhinbiswas/ghost/internal/agent/tools"
	"github.com/swadhinbiswas/ghost/internal/collab"
	"github.com/swadhinbiswas/ghost/internal/config"
	"github.com/swadhinbiswas/ghost/internal/feedback"
	"github.com/swadhinbiswas/ghost/internal/filetracker"
	"github.com/swadhinbiswas/ghost/internal/filewatcher"
	"github.com/swadhinbiswas/ghost/internal/history"
	"github.com/swadhinbiswas/ghost/internal/log"
	"github.com/swadhinbiswas/ghost/internal/lsp"
	"github.com/swadhinbiswas/ghost/internal/memory"
	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/oauth/copilot"
	"github.com/swadhinbiswas/ghost/internal/permission"
	"github.com/swadhinbiswas/ghost/internal/plugin"
	"github.com/swadhinbiswas/ghost/internal/pubsub"
	"github.com/swadhinbiswas/ghost/internal/semantic"
	"github.com/swadhinbiswas/ghost/internal/session"
	"golang.org/x/sync/errgroup"

	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/azure"
	"charm.land/fantasy/providers/bedrock"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
	"charm.land/fantasy/providers/vercel"
	openaisdk "github.com/charmbracelet/openai-go/option"
	"github.com/qjebbs/go-jsons"
)

// Coordinator errors.
var (
	errCoderAgentNotConfigured         = errors.New("coder agent not configured")
	errModelProviderNotConfigured      = errors.New("model provider not configured")
	errLargeModelNotSelected           = errors.New("large model not selected")
	errSmallModelNotSelected           = errors.New("small model not selected")
	errLargeModelProviderNotConfigured = errors.New("large model provider not configured")
	errSmallModelProviderNotConfigured = errors.New("small model provider not configured")
	errLargeModelNotFound              = errors.New("large model not found in provider config")
	errSmallModelNotFound              = errors.New("small model not found in provider config")
)

type Coordinator interface {
	Run(ctx context.Context, sessionID, prompt string, attachments ...message.Attachment) (*fantasy.AgentResult, error)
	Cancel(sessionID string)
	CancelAll()
	SetPaneController(PaneController)
	UpdateModels(ctx context.Context) error
	IsSessionBusy(sessionID string) bool
	IsBusy() bool
	QueuedPrompts(sessionID string) int
	QueuedPromptsList(sessionID string) []string
	ClearQueue(sessionID string)
	Summarize(ctx context.Context, sessionID string) error
	Model() Model
	CollabHub() *collab.Hub
}

// PaneController is the interface for controlling the terminal multiplexer from the agent layer.
type PaneController interface {
	CreatePane(title, content string) string
	UpdateContent(id, content string)
	AppendContent(id, content string)
	ClosePane(id string)
	PaneCount() int
}

type coordinator struct {
	cfg           *config.ConfigStore
	sessions      session.Service
	messages      message.Service
	permissions   permission.Service
	history       history.Service
	filetracker   filetracker.Service
	lspManager    *lsp.Manager
	notify        pubsub.Publisher[notify.Notification]
	filewatcher   *filewatcher.Watcher
	semIndex      *semantic.Index
	memExtractor  *memory.Extractor
	pluginLoader  *plugin.Loader
	feedbackStore *feedback.Store
	collabHub     *collab.Hub
	paneCtrl      PaneController

	currentAgent SessionAgent
	agents       map[string]SessionAgent

	readyWg errgroup.Group
}

func NewCoordinator(
	ctx context.Context,
	cfg *config.ConfigStore,
	sessions session.Service,
	messages message.Service,
	permissions permission.Service,
	history history.Service,
	filetracker filetracker.Service,
	lspManager *lsp.Manager,
	notify pubsub.Publisher[notify.Notification],
) (Coordinator, error) {
	c := &coordinator{
		cfg:           cfg,
		sessions:      sessions,
		messages:      messages,
		permissions:   permissions,
		history:       history,
		filetracker:   filetracker,
		lspManager:    lspManager,
		notify:        notify,
		agents:        make(map[string]SessionAgent),
		filewatcher:   filewatcher.New(cfg.WorkingDir()),
		semIndex:      semantic.NewIndex(cfg.WorkingDir()),
		memExtractor:  memory.NewExtractor(),
		pluginLoader:  plugin.NewLoader(cfg.WorkingDir()),
		feedbackStore: initFeedbackStore(cfg.WorkingDir()),
		collabHub:     collab.NewHub(),
	}

	// Start file watcher to detect external changes
	if c.filewatcher != nil {
		go c.watchExternalChanges(ctx)
	}

	// Build semantic index in background
	go c.buildSemanticIndex(ctx)

	// Load plugins in background
	go c.loadPlugins()

	agentCfg, ok := cfg.Config().Agents[config.AgentCoder]
	if !ok {
		return nil, errCoderAgentNotConfigured
	}

	// TODO: make this dynamic when we support multiple agents
	// Use plan prompt if plan mode is enabled
	var sysPrompt *prompt.Prompt
	var err error
	if cfg.Config().Options.TUI != nil && cfg.Config().Options.TUI.PlanMode {
		sysPrompt, err = planPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))
	} else {
		sysPrompt, err = coderPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))
	}
	if err != nil {
		return nil, err
	}

	agent, err := c.buildAgent(ctx, sysPrompt, agentCfg, false)
	if err != nil {
		return nil, err
	}
	c.currentAgent = agent
	c.agents[config.AgentCoder] = agent
	return c, nil
}

// watchExternalChanges starts the file watcher and handles external file change events.
// When a file the agent has read is modified externally, it tracks the modification
// so the agent can detect stale reads.
func (c *coordinator) watchExternalChanges(ctx context.Context) {
	if c.filewatcher == nil {
		return
	}

	if err := c.filewatcher.Start(ctx); err != nil {
		slog.Warn("Failed to start file watcher", "error", err)
		return
	}

	slog.Info("External file watcher enabled for", "dir", c.cfg.WorkingDir())
}

// SetPaneController sets the multiplexer controller from the UI layer.
func (c *coordinator) SetPaneController(ctrl PaneController) {
	c.paneCtrl = ctrl
}

// CollabHub returns the collaboration WebSocket hub.
func (c *coordinator) CollabHub() *collab.Hub {
	return c.collabHub
}

// buildSemanticIndex builds the semantic search index in the background.
func (c *coordinator) buildSemanticIndex(ctx context.Context) {
	if c.semIndex == nil {
		return
	}

	slog.Info("Building semantic index...", "dir", c.cfg.WorkingDir())
	start := time.Now()
	if err := c.semIndex.Build(ctx); err != nil {
		slog.Warn("Failed to build semantic index", "error", err)
		return
	}
	slog.Info("Semantic index built",
		"chunks", c.semIndex.ChunkCount(),
		"duration", time.Since(start).Round(time.Second),
	)
}

// WasFileModifiedExternally checks if a file was modified externally after a given time.
// Returns true if the file has been changed outside of the agent's control.
func (c *coordinator) WasFileModifiedExternally(path string, since time.Time) bool {
	if c.filewatcher == nil {
		return false
	}
	return c.filewatcher.WasModifiedExternally(path, since)
}

// Run implements Coordinator.
func (c *coordinator) Run(ctx context.Context, sessionID string, prompt string, attachments ...message.Attachment) (*fantasy.AgentResult, error) {
	if err := c.readyWg.Wait(); err != nil {
		return nil, err
	}

	// refresh models before each run
	if err := c.UpdateModels(ctx); err != nil {
		return nil, fmt.Errorf("failed to update models: %w", err)
	}

	model := c.currentAgent.Model()
	maxTokens := model.CatwalkCfg.DefaultMaxTokens
	if model.ModelCfg.MaxTokens != 0 {
		maxTokens = model.ModelCfg.MaxTokens
	}

	if !model.CatwalkCfg.SupportsImages && attachments != nil {
		// filter out image attachments
		filteredAttachments := make([]message.Attachment, 0, len(attachments))
		for _, att := range attachments {
			if att.IsText() {
				filteredAttachments = append(filteredAttachments, att)
			}
		}
		attachments = filteredAttachments
	}

	providerCfg, ok := c.cfg.Config().Providers.Get(model.ModelCfg.Provider)
	if !ok {
		return nil, errModelProviderNotConfigured
	}

	mergedOptions, temp, topP, topK, freqPenalty, presPenalty := mergeCallOptions(model, providerCfg)

	if providerCfg.OAuthToken != nil && providerCfg.OAuthToken.IsExpired() {
		slog.Debug("Token needs to be refreshed", "provider", providerCfg.ID)
		if err := c.refreshOAuth2Token(ctx, providerCfg); err != nil {
			return nil, err
		}
	}

	run := func() (*fantasy.AgentResult, error) {
		return c.currentAgent.Run(ctx, SessionAgentCall{
			SessionID:        sessionID,
			Prompt:           prompt,
			Attachments:      attachments,
			MaxOutputTokens:  maxTokens,
			ProviderOptions:  mergedOptions,
			Temperature:      temp,
			TopP:             topP,
			TopK:             topK,
			FrequencyPenalty: freqPenalty,
			PresencePenalty:  presPenalty,
		})
	}
	result, originalErr := run()

	if originalErr == nil && result != nil {
		go c.extractAndSaveMemories(ctx, sessionID, prompt, result)
	}

	if c.isUnauthorized(originalErr) {
		switch {
		case providerCfg.OAuthToken != nil:
			slog.Debug("Received 401. Refreshing token and retrying", "provider", providerCfg.ID)
			if err := c.refreshOAuth2Token(ctx, providerCfg); err != nil {
				return nil, originalErr
			}
			slog.Debug("Retrying request with refreshed OAuth token", "provider", providerCfg.ID)
			return run()
		case strings.Contains(providerCfg.APIKeyTemplate, "$"):
			slog.Debug("Received 401. Refreshing API Key template and retrying", "provider", providerCfg.ID)
			if err := c.refreshApiKeyTemplate(ctx, providerCfg); err != nil {
				return nil, originalErr
			}
			slog.Debug("Retrying request with refreshed API key", "provider", providerCfg.ID)
			return run()
		}
	}

	return result, originalErr
}

func getProviderOptions(model Model, providerCfg config.ProviderConfig) fantasy.ProviderOptions {
	options := fantasy.ProviderOptions{}

	cfgOpts := []byte("{}")
	providerCfgOpts := []byte("{}")
	catwalkOpts := []byte("{}")

	if model.ModelCfg.ProviderOptions != nil {
		data, err := json.Marshal(model.ModelCfg.ProviderOptions)
		if err == nil {
			cfgOpts = data
		}
	}

	if providerCfg.ProviderOptions != nil {
		data, err := json.Marshal(providerCfg.ProviderOptions)
		if err == nil {
			providerCfgOpts = data
		}
	}

	if model.CatwalkCfg.Options.ProviderOptions != nil {
		data, err := json.Marshal(model.CatwalkCfg.Options.ProviderOptions)
		if err == nil {
			catwalkOpts = data
		}
	}

	readers := []io.Reader{
		bytes.NewReader(catwalkOpts),
		bytes.NewReader(providerCfgOpts),
		bytes.NewReader(cfgOpts),
	}

	got, err := jsons.Merge(readers)
	if err != nil {
		slog.Error("Could not merge call config", "err", err)
		return options
	}

	mergedOptions := make(map[string]any)

	err = json.Unmarshal(got, &mergedOptions)
	if err != nil {
		slog.Error("Could not create config for call", "err", err)
		return options
	}

	providerType := providerCfg.Type
	if providerType == "hyper" {
		if strings.Contains(model.CatwalkCfg.ID, "claude") {
			providerType = anthropic.Name
		} else if strings.Contains(model.CatwalkCfg.ID, "gpt") {
			providerType = openai.Name
		} else if strings.Contains(model.CatwalkCfg.ID, "gemini") {
			providerType = google.Name
		} else {
			providerType = openaicompat.Name
		}
	}

	switch providerType {
	case openai.Name, azure.Name:
		_, hasReasoningEffort := mergedOptions["reasoning_effort"]
		if !hasReasoningEffort && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning_effort"] = model.ModelCfg.ReasoningEffort
		}
		if openai.IsResponsesModel(model.CatwalkCfg.ID) {
			if openai.IsResponsesReasoningModel(model.CatwalkCfg.ID) {
				mergedOptions["reasoning_summary"] = "auto"
				mergedOptions["include"] = []openai.IncludeType{openai.IncludeReasoningEncryptedContent}
			}
			parsed, err := openai.ParseResponsesOptions(mergedOptions)
			if err == nil {
				options[openai.Name] = parsed
			}
		} else {
			parsed, err := openai.ParseOptions(mergedOptions)
			if err == nil {
				options[openai.Name] = parsed
			}
		}
	case anthropic.Name:
		var (
			_, hasEffort = mergedOptions["effort"]
			_, hasThink  = mergedOptions["thinking"]
		)
		switch {
		case !hasEffort && model.ModelCfg.ReasoningEffort != "":
			mergedOptions["effort"] = model.ModelCfg.ReasoningEffort
		case !hasThink && model.ModelCfg.Think:
			mergedOptions["thinking"] = map[string]any{"budget_tokens": 2000}
		}
		parsed, err := anthropic.ParseOptions(mergedOptions)
		if err == nil {
			options[anthropic.Name] = parsed
		}

	case openrouter.Name:
		_, hasReasoning := mergedOptions["reasoning"]
		if !hasReasoning && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning"] = map[string]any{
				"enabled": true,
				"effort":  model.ModelCfg.ReasoningEffort,
			}
		}
		parsed, err := openrouter.ParseOptions(mergedOptions)
		if err == nil {
			options[openrouter.Name] = parsed
		}
	case vercel.Name:
		_, hasReasoning := mergedOptions["reasoning"]
		if !hasReasoning && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning"] = map[string]any{
				"enabled": true,
				"effort":  model.ModelCfg.ReasoningEffort,
			}
		}
		parsed, err := vercel.ParseOptions(mergedOptions)
		if err == nil {
			options[vercel.Name] = parsed
		}
	case google.Name:
		_, hasReasoning := mergedOptions["thinking_config"]
		if !hasReasoning {
			if strings.HasPrefix(model.CatwalkCfg.ID, "gemini-2") {
				mergedOptions["thinking_config"] = map[string]any{
					"thinking_budget":  2000,
					"include_thoughts": true,
				}
			} else {
				mergedOptions["thinking_config"] = map[string]any{
					"thinking_level":   model.ModelCfg.ReasoningEffort,
					"include_thoughts": true,
				}
			}
		}
		parsed, err := google.ParseOptions(mergedOptions)
		if err == nil {
			options[google.Name] = parsed
		}
	case openaicompat.Name:
		_, hasReasoningEffort := mergedOptions["reasoning_effort"]
		if !hasReasoningEffort && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning_effort"] = model.ModelCfg.ReasoningEffort
		}
		parsed, err := openaicompat.ParseOptions(mergedOptions)
		if err == nil {
			options[openaicompat.Name] = parsed
		}
	}

	return options
}

func mergeCallOptions(model Model, cfg config.ProviderConfig) (fantasy.ProviderOptions, *float64, *float64, *int64, *float64, *float64) {
	modelOptions := getProviderOptions(model, cfg)
	temp := cmp.Or(model.ModelCfg.Temperature, model.CatwalkCfg.Options.Temperature)
	topP := cmp.Or(model.ModelCfg.TopP, model.CatwalkCfg.Options.TopP)
	topK := cmp.Or(model.ModelCfg.TopK, model.CatwalkCfg.Options.TopK)
	freqPenalty := cmp.Or(model.ModelCfg.FrequencyPenalty, model.CatwalkCfg.Options.FrequencyPenalty)
	presPenalty := cmp.Or(model.ModelCfg.PresencePenalty, model.CatwalkCfg.Options.PresencePenalty)
	return modelOptions, temp, topP, topK, freqPenalty, presPenalty
}

func (c *coordinator) buildAgent(ctx context.Context, prompt *prompt.Prompt, agent config.Agent, isSubAgent bool) (SessionAgent, error) {
	large, small, err := c.buildAgentModels(ctx, isSubAgent)
	if err != nil {
		return nil, err
	}

	largeProviderCfg, _ := c.cfg.Config().Providers.Get(large.ModelCfg.Provider)
	result := NewSessionAgent(SessionAgentOptions{
		LargeModel:           large,
		SmallModel:           small,
		SystemPromptPrefix:   largeProviderCfg.SystemPromptPrefix,
		SystemPrompt:         "",
		IsSubAgent:           isSubAgent,
		DisableAutoSummarize: c.cfg.Config().Options.DisableAutoSummarize,
		IsYolo:               c.permissions.SkipRequests(),
		Sessions:             c.sessions,
		Messages:             c.messages,
		Tools:                nil,
		Notify:               c.notify,
		Config:               c.cfg,
	})

	c.readyWg.Go(func() error {
		systemPrompt, err := prompt.Build(ctx, large.Model.Provider(), large.Model.Model(), c.cfg)
		if err != nil {
			return err
		}
		// Append self-healing instructions to enable automatic error recovery
		systemPrompt += SelfHealingPrompt
		result.SetSystemPrompt(systemPrompt)
		return nil
	})

	c.readyWg.Go(func() error {
		tools, err := c.buildTools(ctx, agent)
		if err != nil {
			return err
		}
		result.SetTools(tools)
		return nil
	})

	return result, nil
}

func (c *coordinator) buildTools(ctx context.Context, agent config.Agent) ([]fantasy.AgentTool, error) {
	var allTools []fantasy.AgentTool
	if slices.Contains(agent.AllowedTools, AgentToolName) {
		agentTool, err := c.agentTool(ctx)
		if err != nil {
			return nil, err
		}
		allTools = append(allTools, agentTool)
	}

	if slices.Contains(agent.AllowedTools, tools.AgenticFetchToolName) {
		agenticFetchTool, err := c.agenticFetchTool(ctx, nil)
		if err != nil {
			return nil, err
		}
		allTools = append(allTools, agenticFetchTool)
	}

	// Get the model name for the agent
	modelName := ""
	if modelCfg, ok := c.cfg.Config().Models[agent.Model]; ok {
		if model := c.cfg.Config().GetModel(modelCfg.Provider, modelCfg.Model); model != nil {
			modelName = model.Name
		}
	}

	// Register plugin tools
	pluginTools := c.buildPluginTools()
	allTools = append(allTools, pluginTools...)

	// Register plugin agents
	pluginAgents := c.buildPluginAgents(ctx)
	allTools = append(allTools, pluginAgents...)

	allTools = append(allTools,
		tools.NewFetchTool(c.permissions, c.cfg.WorkingDir(), nil),
		tools.NewGlobTool(c.cfg.WorkingDir()),
		tools.NewGrepTool(c.cfg.WorkingDir(), c.cfg.Config().Tools.Grep),
		tools.NewLsTool(c.permissions, c.cfg.WorkingDir(), c.cfg.Config().Tools.Ls),
		tools.NewSemanticSearchTool(c.semIndex, c.cfg.WorkingDir()),
		tools.NewSourcegraphTool(nil),
		tools.NewViewTool(c.lspManager, c.permissions, c.filetracker, c.cfg.WorkingDir(), c.cfg.Config().Options.SkillsPaths...),
		tools.NewSymbolsTool(c.cfg.WorkingDir()),
	)

	// Plan mode: read-only tools only
	planMode := c.cfg.Config().Options.TUI != nil && c.cfg.Config().Options.TUI.PlanMode
	if !planMode {
		allTools = append(allTools,
			tools.NewBashTool(c.permissions, c.cfg.WorkingDir(), c.cfg.Config().Options, modelName),
			tools.NewJobOutputTool(),
			tools.NewJobKillTool(),
			tools.NewDownloadTool(c.permissions, c.cfg.WorkingDir(), nil),
			tools.NewEditTool(c.lspManager, c.permissions, c.history, c.filetracker, c.cfg.WorkingDir()),
			tools.NewMultiEditTool(c.lspManager, c.permissions, c.history, c.filetracker, c.cfg.WorkingDir()),
			tools.NewAtomicEditTool(c.permissions, c.history, c.cfg.WorkingDir()),
			tools.NewMemoryTool(c.cfg),
			tools.NewTodosTool(c.sessions),
			tools.NewWriteTool(c.lspManager, c.permissions, c.history, c.filetracker, c.cfg.WorkingDir()),
			tools.NewUndoTool(c.permissions, c.history, c.cfg.WorkingDir()),
			tools.NewTestTool(c.permissions, c.cfg.WorkingDir()),
			tools.NewScreenshotTool(c.cfg.WorkingDir()),
			tools.NewSwarmTool(c.swarmRunner(), c.cfg.WorkingDir()),
			tools.NewPluginTool(c.pluginLoader),
			tools.NewFeedbackTool(c.feedbackStore),
			tools.NewCollabTool(c.collabHub),
			tools.NewMultiplexTool(func() tools.PaneManager { return c.paneCtrl }),
		)
	} else {
		// In plan mode, add a system prompt prefix indicating read-only mode
		slog.Info("Plan mode enabled: only read-only tools available")
	}

	// Add LSP tools if user has configured LSPs or auto_lsp is enabled (nil or true).
	if len(c.cfg.Config().LSP) > 0 || c.cfg.Config().Options.AutoLSP == nil || *c.cfg.Config().Options.AutoLSP {
		allTools = append(allTools, tools.NewDiagnosticsTool(c.lspManager), tools.NewReferencesTool(c.lspManager), tools.NewLSPRestartTool(c.lspManager))
	}

	if len(c.cfg.Config().MCP) > 0 {
		allTools = append(
			allTools,
			tools.NewListMCPResourcesTool(c.cfg, c.permissions),
			tools.NewReadMCPResourceTool(c.cfg, c.permissions),
		)
	}

	var filteredTools []fantasy.AgentTool
	for _, tool := range allTools {
		if slices.Contains(agent.AllowedTools, tool.Info().Name) {
			filteredTools = append(filteredTools, tool)
		}
	}

	for _, tool := range tools.GetMCPTools(c.permissions, c.cfg, c.cfg.WorkingDir()) {
		if agent.AllowedMCP == nil {
			// No MCP restrictions
			filteredTools = append(filteredTools, tool)
			continue
		}
		if len(agent.AllowedMCP) == 0 {
			// No MCPs allowed
			slog.Debug("No MCPs allowed", "tool", tool.Name(), "agent", agent.Name)
			break
		}

		for mcp, tools := range agent.AllowedMCP {
			if mcp != tool.MCP() {
				continue
			}
			if len(tools) == 0 || slices.Contains(tools, tool.MCPToolName()) {
				filteredTools = append(filteredTools, tool)
				break
			}
			slog.Debug("MCP not allowed", "tool", tool.Name(), "agent", agent.Name)
		}
	}
	slices.SortFunc(filteredTools, func(a, b fantasy.AgentTool) int {
		return strings.Compare(a.Info().Name, b.Info().Name)
	})
	return filteredTools, nil
}

// TODO: when we support multiple agents we need to change this so that we pass in the agent specific model config
func (c *coordinator) buildAgentModels(ctx context.Context, isSubAgent bool) (Model, Model, error) {
	largeModelCfg, ok := c.cfg.Config().Models[config.SelectedModelTypeLarge]
	if !ok {
		return Model{}, Model{}, errLargeModelNotSelected
	}
	smallModelCfg, ok := c.cfg.Config().Models[config.SelectedModelTypeSmall]
	if !ok {
		return Model{}, Model{}, errSmallModelNotSelected
	}

	largeProviderCfg, ok := c.cfg.Config().Providers.Get(largeModelCfg.Provider)
	if !ok {
		return Model{}, Model{}, errLargeModelProviderNotConfigured
	}

	largeProvider, err := c.buildProvider(largeProviderCfg, largeModelCfg, isSubAgent)
	if err != nil {
		return Model{}, Model{}, err
	}

	smallProviderCfg, ok := c.cfg.Config().Providers.Get(smallModelCfg.Provider)
	if !ok {
		return Model{}, Model{}, errSmallModelProviderNotConfigured
	}

	smallProvider, err := c.buildProvider(smallProviderCfg, smallModelCfg, true)
	if err != nil {
		return Model{}, Model{}, err
	}

	var largeCatwalkModel *catwalk.Model
	var smallCatwalkModel *catwalk.Model

	for _, m := range largeProviderCfg.Models {
		if m.ID == largeModelCfg.Model {
			largeCatwalkModel = &m
		}
	}
	for _, m := range smallProviderCfg.Models {
		if m.ID == smallModelCfg.Model {
			smallCatwalkModel = &m
		}
	}

	if largeCatwalkModel == nil {
		return Model{}, Model{}, errLargeModelNotFound
	}

	if smallCatwalkModel == nil {
		return Model{}, Model{}, errSmallModelNotFound
	}

	largeModelID := largeModelCfg.Model
	smallModelID := smallModelCfg.Model

	if largeModelCfg.Provider == openrouter.Name && isExactoSupported(largeModelID) {
		largeModelID += ":exacto"
	}

	if smallModelCfg.Provider == openrouter.Name && isExactoSupported(smallModelID) {
		smallModelID += ":exacto"
	}

	largeModel, err := largeProvider.LanguageModel(ctx, largeModelID)
	if err != nil {
		return Model{}, Model{}, err
	}
	smallModel, err := smallProvider.LanguageModel(ctx, smallModelID)
	if err != nil {
		return Model{}, Model{}, err
	}

	return Model{
			Model:      largeModel,
			CatwalkCfg: *largeCatwalkModel,
			ModelCfg:   largeModelCfg,
		}, Model{
			Model:      smallModel,
			CatwalkCfg: *smallCatwalkModel,
			ModelCfg:   smallModelCfg,
		}, nil
}

func (c *coordinator) buildAnthropicProvider(baseURL, apiKey string, headers map[string]string, providerID string) (fantasy.Provider, error) {
	var opts []anthropic.Option

	switch {
	case strings.HasPrefix(apiKey, "Bearer "):
		// NOTE: Prevent the SDK from picking up the API key from env.
		os.Setenv("ANTHROPIC_API_KEY", "")
		headers["Authorization"] = apiKey
	case providerID == string(catwalk.InferenceProviderMiniMax) || providerID == string(catwalk.InferenceProviderMiniMaxChina):
		// NOTE: Prevent the SDK from picking up the API key from env.
		os.Setenv("ANTHROPIC_API_KEY", "")
		headers["Authorization"] = "Bearer " + apiKey
	case apiKey != "":
		// X-Api-Key header
		opts = append(opts, anthropic.WithAPIKey(apiKey))
	}

	if len(headers) > 0 {
		opts = append(opts, anthropic.WithHeaders(headers))
	}

	if baseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(baseURL))
	}

	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, anthropic.WithHTTPClient(httpClient))
	}
	return anthropic.New(opts...)
}

func (c *coordinator) buildOpenaiProvider(baseURL, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []openai.Option{
		openai.WithUseResponsesAPI(),
	}
	if apiKey != "" {
		opts = append(opts, openai.WithAPIKey(apiKey))
	}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, openai.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, openai.WithHeaders(headers))
	}
	if baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}
	return openai.New(opts...)
}

func (c *coordinator) buildOpenrouterProvider(_, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []openrouter.Option{
		openrouter.WithAPIKey(apiKey),
	}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, openrouter.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, openrouter.WithHeaders(headers))
	}
	return openrouter.New(opts...)
}

func (c *coordinator) buildVercelProvider(_, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []vercel.Option{
		vercel.WithAPIKey(apiKey),
	}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, vercel.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, vercel.WithHeaders(headers))
	}
	return vercel.New(opts...)
}

func (c *coordinator) buildOpenaiCompatProvider(baseURL, apiKey string, headers map[string]string, extraBody map[string]any, providerID string, isSubAgent bool) (fantasy.Provider, error) {
	opts := []openaicompat.Option{
		openaicompat.WithBaseURL(baseURL),
	}
	// Set HTTP client based on provider and debug mode.
	var httpClient *http.Client
	if providerID == string(catwalk.InferenceProviderCopilot) {
		opts = append(opts, openaicompat.WithUseResponsesAPI())
		httpClient = copilot.NewClient(isSubAgent, c.cfg.Config().Options.Debug)
	} else if c.cfg.Config().Options.Debug {
		httpClient = log.NewHTTPClient()
	} else {
		httpClient = &http.Client{}
	}

	if apiKey == "no-key-needed" {
		if httpClient.Transport == nil {
			httpClient.Transport = http.DefaultTransport
		}
		httpClient.Transport = &headerStrippingTransport{
			Transport: httpClient.Transport,
			Headers:   []string{"Authorization"},
		}
		// Reset apiKey so we don't pass 'no-key-needed' literally
		apiKey = ""
	}

	if apiKey != "" {
		opts = append(opts, openaicompat.WithAPIKey(apiKey))
	}

	if httpClient != nil {
		opts = append(opts, openaicompat.WithHTTPClient(httpClient))
	}

	if len(headers) > 0 {
		opts = append(opts, openaicompat.WithHeaders(headers))
	}

	for extraKey, extraValue := range extraBody {
		opts = append(opts, openaicompat.WithSDKOptions(openaisdk.WithJSONSet(extraKey, extraValue)))
	}

	return openaicompat.New(opts...)
}

func (c *coordinator) buildAzureProvider(baseURL, apiKey string, headers map[string]string, options map[string]string) (fantasy.Provider, error) {
	opts := []azure.Option{
		azure.WithBaseURL(baseURL),
		azure.WithAPIKey(apiKey),
		azure.WithUseResponsesAPI(),
	}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, azure.WithHTTPClient(httpClient))
	}
	if options == nil {
		options = make(map[string]string)
	}
	if apiVersion, ok := options["apiVersion"]; ok {
		opts = append(opts, azure.WithAPIVersion(apiVersion))
	}
	if len(headers) > 0 {
		opts = append(opts, azure.WithHeaders(headers))
	}

	return azure.New(opts...)
}

func (c *coordinator) buildBedrockProvider(apiKey string, headers map[string]string) (fantasy.Provider, error) {
	var opts []bedrock.Option
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, bedrock.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, bedrock.WithHeaders(headers))
	}
	switch {
	case apiKey != "":
		opts = append(opts, bedrock.WithAPIKey(apiKey))
	case os.Getenv("AWS_BEARER_TOKEN_BEDROCK") != "":
		opts = append(opts, bedrock.WithAPIKey(os.Getenv("AWS_BEARER_TOKEN_BEDROCK")))
	default:
		// Skip, let the SDK do authentication.
	}
	return bedrock.New(opts...)
}

func (c *coordinator) buildGoogleProvider(baseURL, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []google.Option{
		google.WithBaseURL(baseURL),
		google.WithGeminiAPIKey(apiKey),
	}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, google.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, google.WithHeaders(headers))
	}
	return google.New(opts...)
}

func (c *coordinator) buildGoogleVertexProvider(headers map[string]string, options map[string]string) (fantasy.Provider, error) {
	opts := []google.Option{}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, google.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, google.WithHeaders(headers))
	}

	project := options["project"]
	location := options["location"]

	opts = append(opts, google.WithVertex(project, location))

	return google.New(opts...)
}

func (c *coordinator) buildHyperProvider(apiKey string) (fantasy.Provider, error) {
	opts := []hyper.Option{
		hyper.WithAPIKey(apiKey),
	}
	if c.cfg.Config().Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, hyper.WithHTTPClient(httpClient))
	}
	return hyper.New(opts...)
}

func (c *coordinator) isAnthropicThinking(model config.SelectedModel) bool {
	if model.Think {
		return true
	}
	opts, err := anthropic.ParseOptions(model.ProviderOptions)
	return err == nil && opts.Thinking != nil
}

func (c *coordinator) buildProvider(providerCfg config.ProviderConfig, model config.SelectedModel, isSubAgent bool) (fantasy.Provider, error) {
	headers := maps.Clone(providerCfg.ExtraHeaders)
	if headers == nil {
		headers = make(map[string]string)
	}

	// handle special headers for anthropic
	if providerCfg.Type == anthropic.Name && c.isAnthropicThinking(model) {
		if v, ok := headers["anthropic-beta"]; ok {
			headers["anthropic-beta"] = v + ",interleaved-thinking-2025-05-14"
		} else {
			headers["anthropic-beta"] = "interleaved-thinking-2025-05-14"
		}
	}

	apiKey, _ := c.cfg.Resolve(providerCfg.APIKey)
	baseURL, _ := c.cfg.Resolve(providerCfg.BaseURL)

	switch providerCfg.Type {
	case openai.Name:
		return c.buildOpenaiProvider(baseURL, apiKey, headers)
	case anthropic.Name:
		return c.buildAnthropicProvider(baseURL, apiKey, headers, providerCfg.ID)
	case openrouter.Name:
		return c.buildOpenrouterProvider(baseURL, apiKey, headers)
	case vercel.Name:
		return c.buildVercelProvider(baseURL, apiKey, headers)
	case azure.Name:
		return c.buildAzureProvider(baseURL, apiKey, headers, providerCfg.ExtraParams)
	case bedrock.Name:
		return c.buildBedrockProvider(apiKey, headers)
	case google.Name:
		return c.buildGoogleProvider(baseURL, apiKey, headers)
	case "google-vertex":
		return c.buildGoogleVertexProvider(headers, providerCfg.ExtraParams)
	case openaicompat.Name:
		if providerCfg.ID == string(catwalk.InferenceProviderZAI) {
			if providerCfg.ExtraBody == nil {
				providerCfg.ExtraBody = map[string]any{}
			}
			providerCfg.ExtraBody["tool_stream"] = true
		}
		if providerCfg.ID == "opencode" || providerCfg.ID == "opencode-zen" {
			headers["User-Agent"] = "OpenCode/1.0.0"
			headers["X-Client-Name"] = "OpenCode"
			if providerCfg.ExtraBody == nil {
				providerCfg.ExtraBody = map[string]any{}
			}
			switch model.Model {
			case "nemotron-3-super-free":
				providerCfg.ExtraBody["chat_template_kwargs"] = map[string]any{"enable_thinking": true}
				providerCfg.ExtraBody["reasoning_budget"] = 16384
			case "deepseek-v4-flash-free", "big-pickle":
				providerCfg.ExtraBody["chat_template_kwargs"] = map[string]any{"thinking": true, "reasoning_effort": "high"}
			}
		}
		if providerCfg.ID == "nvidia-nim" {
			if providerCfg.ExtraBody == nil {
				providerCfg.ExtraBody = map[string]any{}
			}
			switch model.Model {
			case "nvidia/nemotron-3-ultra-550b-a55b", "nvidia/nemotron-3-super-120b-a12b":
				providerCfg.ExtraBody["chat_template_kwargs"] = map[string]any{"enable_thinking": true}
				providerCfg.ExtraBody["reasoning_budget"] = 16384
			case "deepseek-ai/deepseek-v4-flash":
				providerCfg.ExtraBody["chat_template_kwargs"] = map[string]any{"thinking": true, "reasoning_effort": "high"}
			case "google/gemma-4-31b-it":
				providerCfg.ExtraBody["chat_template_kwargs"] = map[string]any{"enable_thinking": true}
			}
		}
		return c.buildOpenaiCompatProvider(baseURL, apiKey, headers, providerCfg.ExtraBody, providerCfg.ID, isSubAgent)
	case hyper.Name:
		return c.buildHyperProvider(apiKey)
	default:
		return nil, fmt.Errorf("provider type not supported: %q", providerCfg.Type)
	}
}

func isExactoSupported(modelID string) bool {
	supportedModels := []string{
		"moonshotai/kimi-k2-0905",
		"deepseek/deepseek-v3.1-terminus",
		"z-ai/glm-4.6",
		"openai/gpt-oss-120b",
		"qwen/qwen3-coder",
	}
	return slices.Contains(supportedModels, modelID)
}

func (c *coordinator) Cancel(sessionID string) {
	c.currentAgent.Cancel(sessionID)
}

func (c *coordinator) CancelAll() {
	c.currentAgent.CancelAll()
}

func (c *coordinator) ClearQueue(sessionID string) {
	c.currentAgent.ClearQueue(sessionID)
}

func (c *coordinator) IsBusy() bool {
	return c.currentAgent.IsBusy()
}

func (c *coordinator) IsSessionBusy(sessionID string) bool {
	return c.currentAgent.IsSessionBusy(sessionID)
}

func (c *coordinator) Model() Model {
	return c.currentAgent.Model()
}

func (c *coordinator) UpdateModels(ctx context.Context) error {
	// build the models again so we make sure we get the latest config
	large, small, err := c.buildAgentModels(ctx, false)
	if err != nil {
		return err
	}
	c.currentAgent.SetModels(large, small)

	agentCfg, ok := c.cfg.Config().Agents[config.AgentCoder]
	if !ok {
		return errCoderAgentNotConfigured
	}

	tools, err := c.buildTools(ctx, agentCfg)
	if err != nil {
		return err
	}
	c.currentAgent.SetTools(tools)
	return nil
}

func (c *coordinator) QueuedPrompts(sessionID string) int {
	return c.currentAgent.QueuedPrompts(sessionID)
}

func (c *coordinator) QueuedPromptsList(sessionID string) []string {
	return c.currentAgent.QueuedPromptsList(sessionID)
}

func (c *coordinator) Summarize(ctx context.Context, sessionID string) error {
	providerCfg, ok := c.cfg.Config().Providers.Get(c.currentAgent.Model().ModelCfg.Provider)
	if !ok {
		return errModelProviderNotConfigured
	}
	return c.currentAgent.Summarize(ctx, sessionID, getProviderOptions(c.currentAgent.Model(), providerCfg))
}

func (c *coordinator) isUnauthorized(err error) bool {
	var providerErr *fantasy.ProviderError
	return errors.As(err, &providerErr) && providerErr.StatusCode == http.StatusUnauthorized
}

func (c *coordinator) refreshOAuth2Token(ctx context.Context, providerCfg config.ProviderConfig) error {
	if err := c.cfg.RefreshOAuthToken(ctx, config.ScopeGlobal, providerCfg.ID); err != nil {
		slog.Error("Failed to refresh OAuth token after 401 error", "provider", providerCfg.ID, "error", err)
		return err
	}
	if err := c.UpdateModels(ctx); err != nil {
		return err
	}
	return nil
}

func (c *coordinator) refreshApiKeyTemplate(ctx context.Context, providerCfg config.ProviderConfig) error {
	newAPIKey, err := c.cfg.Resolve(providerCfg.APIKeyTemplate)
	if err != nil {
		slog.Error("Failed to re-resolve API key after 401 error", "provider", providerCfg.ID, "error", err)
		return err
	}

	providerCfg.APIKey = newAPIKey
	c.cfg.Config().Providers.Set(providerCfg.ID, providerCfg)

	if err := c.UpdateModels(ctx); err != nil {
		return err
	}
	return nil
}

// subAgentParams holds the parameters for running a sub-agent.
type subAgentParams struct {
	Agent          SessionAgent
	SessionID      string
	AgentMessageID string
	ToolCallID     string
	Prompt         string
	SessionTitle   string
	// SessionSetup is an optional callback invoked after session creation
	// but before agent execution, for custom session configuration.
	SessionSetup func(sessionID string)
}

// runSubAgent runs a sub-agent and handles session management and cost accumulation.
// It creates a sub-session, runs the agent with the given prompt, and propagates
// the cost to the parent session.
func (c *coordinator) runSubAgent(ctx context.Context, params subAgentParams) (fantasy.ToolResponse, error) {
	// Create sub-session
	agentToolSessionID := c.sessions.CreateAgentToolSessionID(params.AgentMessageID, params.ToolCallID)
	session, err := c.sessions.CreateTaskSession(ctx, agentToolSessionID, params.SessionID, params.SessionTitle)
	if err != nil {
		return fantasy.ToolResponse{}, fmt.Errorf("create session: %w", err)
	}

	// Call session setup function if provided
	if params.SessionSetup != nil {
		params.SessionSetup(session.ID)
	}

	// Get model configuration
	model := params.Agent.Model()
	maxTokens := model.CatwalkCfg.DefaultMaxTokens
	if model.ModelCfg.MaxTokens != 0 {
		maxTokens = model.ModelCfg.MaxTokens
	}

	providerCfg, ok := c.cfg.Config().Providers.Get(model.ModelCfg.Provider)
	if !ok {
		return fantasy.ToolResponse{}, errModelProviderNotConfigured
	}

	// Run the agent
	result, err := params.Agent.Run(ctx, SessionAgentCall{
		SessionID:        session.ID,
		Prompt:           params.Prompt,
		MaxOutputTokens:  maxTokens,
		ProviderOptions:  getProviderOptions(model, providerCfg),
		Temperature:      model.ModelCfg.Temperature,
		TopP:             model.ModelCfg.TopP,
		TopK:             model.ModelCfg.TopK,
		FrequencyPenalty: model.ModelCfg.FrequencyPenalty,
		PresencePenalty:  model.ModelCfg.PresencePenalty,
		NonInteractive:   true,
	})
	if err != nil {
		return fantasy.NewTextErrorResponse("error generating response"), nil
	}

	// Update parent session cost
	if err := c.updateParentSessionCost(ctx, session.ID, params.SessionID); err != nil {
		return fantasy.ToolResponse{}, err
	}

	return fantasy.NewTextResponse(result.Response.Content.Text()), nil
}

// updateParentSessionCost accumulates the cost from a child session to its parent session.
func (c *coordinator) updateParentSessionCost(ctx context.Context, childSessionID, parentSessionID string) error {
	childSession, err := c.sessions.Get(ctx, childSessionID)
	if err != nil {
		return fmt.Errorf("get child session: %w", err)
	}

	parentSession, err := c.sessions.Get(ctx, parentSessionID)
	if err != nil {
		return fmt.Errorf("get parent session: %w", err)
	}

	parentSession.Cost += childSession.Cost

	if _, err := c.sessions.Save(ctx, parentSession); err != nil {
		return fmt.Errorf("save parent session: %w", err)
	}

	return nil
}

// extractAndSaveMemories extracts memorable facts from the conversation and saves them.
// This runs asynchronously after a successful agent run to enable cross-session learning.
func (c *coordinator) extractAndSaveMemories(ctx context.Context, sessionID, userPrompt string, result *fantasy.AgentResult) {
	if c.memExtractor == nil {
		return
	}

	messages := c.getConversationMessages(ctx, sessionID)
	if len(messages) == 0 {
		return
	}

	facts := c.memExtractor.ExtractFromConversation(messages)

	if len(result.Response.Content) > 0 {
		responseFacts := memory.ExtractKeyFacts(result.Response.Content.Text())
		facts = append(facts, responseFacts...)
	}

	if len(facts) == 0 {
		return
	}

	memStore, err := memory.NewMemoryStore(c.cfg.WorkingDir())
	if err != nil {
		slog.Debug("Failed to init memory store for extraction", "error", err)
		return
	}

	existing := memStore.GetAll()
	saved := 0
	for _, fact := range facts {
		key := fact.Category + "_" + fact.Key
		if _, exists := existing[key]; exists {
			continue
		}
		value := fact.Value
		if len(value) > 300 {
			value = value[:300] + "..."
		}
		if err := memStore.Set(key, value); err != nil {
			slog.Debug("Failed to save extracted memory", "key", key, "error", err)
			continue
		}
		saved++
	}

	if saved > 0 {
		slog.Info("Extracted and saved memories", "count", saved, "session", sessionID)
	}
}

// getConversationMessages retrieves the conversation messages for memory extraction.
func (c *coordinator) getConversationMessages(ctx context.Context, sessionID string) []memory.ConversationMessage {
	msgs, err := c.messages.List(ctx, sessionID)
	if err != nil {
		return nil
	}

	var result []memory.ConversationMessage
	for _, msg := range msgs {
		role := string(msg.Role)
		content := msg.Content().Text
		if content != "" {
			result = append(result, memory.ConversationMessage{
				Role:    role,
				Content: content,
			})
		}
	}
	return result
}

// swarmRunner returns a SubAgentRunner function for use by the swarm tool.
func (c *coordinator) swarmRunner() tools.SubAgentRunner {
	return func(ctx context.Context, params tools.SubAgentRunParams) (fantasy.ToolResponse, error) {
		return c.runSubAgent(ctx, subAgentParams{
			Agent:          c.currentAgent,
			SessionID:      params.SessionID,
			AgentMessageID: params.AgentMessageID,
			ToolCallID:     params.ToolCallID,
			Prompt:         params.Prompt,
			SessionTitle:   params.SessionTitle,
		})
	}
}

// loadPlugins discovers and loads all plugins in the background.
func (c *coordinator) loadPlugins() {
	if c.pluginLoader == nil {
		return
	}

	plugins, err := c.pluginLoader.LoadAll()
	if err != nil {
		slog.Warn("Failed to load plugins", "error", err)
		return
	}

	if len(plugins) > 0 {
		slog.Info("Plugins loaded", "count", len(plugins))
		for _, p := range plugins {
			slog.Info("Plugin", "name", p.Manifest.Name, "type", p.Manifest.Type, "version", p.Manifest.Version)
		}
	}
}

// initFeedbackStore initializes the feedback store, ignoring errors if the directory doesn't exist yet.
func initFeedbackStore(workingDir string) *feedback.Store {
	store, err := feedback.NewStore(workingDir)
	if err != nil {
		slog.Debug("Feedback store not available", "error", err)
		return nil
	}
	return store
}

// buildPluginTools creates agent tools from loaded tool-type plugins.
func (c *coordinator) buildPluginTools() []fantasy.AgentTool {
	if c.pluginLoader == nil {
		return nil
	}

	toolPlugins := c.pluginLoader.GetTools()
	if len(toolPlugins) == 0 {
		return nil
	}

	var tools []fantasy.AgentTool
	for _, p := range toolPlugins {
		if p.Manifest.ToolDef == nil {
			continue
		}

		td := p.Manifest.ToolDef
		tool := fantasy.NewAgentTool(
			td.Name,
			td.Description,
			func(ctx context.Context, params map[string]interface{}, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				input, _ := json.Marshal(params)
				output, err := p.Runtime.Execute(ctx, string(input))
				if err != nil {
					return fantasy.ToolResponse{Content: output, IsError: true}, nil
				}
				return fantasy.NewTextResponse(output), nil
			},
		)
		tools = append(tools, tool)
		slog.Info("Plugin tool registered", "plugin", p.Manifest.Name, "tool", td.Name)
	}

	return tools
}

// buildPluginAgents creates agent tools from loaded agent-type plugins.
func (c *coordinator) buildPluginAgents(ctx context.Context) []fantasy.AgentTool {
	if c.pluginLoader == nil {
		return nil
	}

	agentPlugins := c.pluginLoader.GetAgents()
	if len(agentPlugins) == 0 {
		return nil
	}

	var agentTools []fantasy.AgentTool
	for _, p := range agentPlugins {
		if p.Manifest.AgentDef == nil {
			continue
		}

		ad := p.Manifest.AgentDef

		// Create agent config
		agentCfg := config.Agent{
			ID:           ad.Name,
			Name:         ad.Name,
			Description:  ad.Description,
			Model:        config.SelectedModelTypeLarge,
			AllowedTools: ad.Tools,
		}
		if strings.ToLower(ad.Model) == "small" {
			agentCfg.Model = config.SelectedModelTypeSmall
		}

		// Create system prompt template
		promptTemplate, err := prompt.NewPrompt(ad.Name, ad.SystemPrompt, prompt.WithWorkingDir(c.cfg.WorkingDir()))
		if err != nil {
			slog.Warn("Failed to create custom agent prompt", "agent", ad.Name, "error", err)
			continue
		}

		// Build agent
		subAgent, err := c.buildAgent(ctx, promptTemplate, agentCfg, true)
		if err != nil {
			slog.Warn("Failed to build custom agent", "agent", ad.Name, "error", err)
			continue
		}

		// Store in our agents map
		c.agents[ad.Name] = subAgent

		// Register as a ParallelAgentTool
		toolName := "agent_" + ad.Name
		toolDescription := ad.Description + "\n\nUse this tool to delegate tasks to the specialized '" + ad.Name + "' agent."

		tool := fantasy.NewParallelAgentTool(
			toolName,
			toolDescription,
			func(ctx context.Context, params AgentParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
				if params.Prompt == "" {
					return fantasy.NewTextErrorResponse("prompt is required"), nil
				}

				sessionID := tools.GetSessionFromContext(ctx)
				if sessionID == "" {
					return fantasy.ToolResponse{}, errors.New("session id missing from context")
				}

				agentMessageID := tools.GetMessageFromContext(ctx)
				if agentMessageID == "" {
					return fantasy.ToolResponse{}, errors.New("agent message id missing from context")
				}

				return c.runSubAgent(ctx, subAgentParams{
					Agent:          subAgent,
					SessionID:      sessionID,
					AgentMessageID: agentMessageID,
					ToolCallID:     call.ID,
					Prompt:         params.Prompt,
					SessionTitle:   fmt.Sprintf("Agent: %s", ad.Name),
				})
			},
		)

		agentTools = append(agentTools, tool)
		slog.Info("Plugin agent registered as tool", "plugin", p.Manifest.Name, "agent", ad.Name, "tool", toolName)
	}

	return agentTools
}

type headerStrippingTransport struct {
	Transport http.RoundTripper
	Headers   []string
}

func (t *headerStrippingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, h := range t.Headers {
		req.Header.Del(h)
	}
	return t.Transport.RoundTrip(req)
}
