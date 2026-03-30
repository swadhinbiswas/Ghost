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

	"charm.land/catwalk/pkg/catwalk"
	"charm.land/fantasy"
	"github.com/charmbracelet/crush/internal/agent/hyper"
	"github.com/charmbracelet/crush/internal/agent/notify"
	"github.com/charmbracelet/crush/internal/agent/prompt"
	"github.com/charmbracelet/crush/internal/agent/tools"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/filetracker"
	"github.com/charmbracelet/crush/internal/history"
	"github.com/charmbracelet/crush/internal/log"
	"github.com/charmbracelet/crush/internal/lsp"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/oauth/copilot"
	"github.com/charmbracelet/crush/internal/permission"
	"github.com/charmbracelet/crush/internal/pubsub"
	"github.com/charmbracelet/crush/internal/session"
	"github.com/esiqveland/notify"
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


type Cordinator interface{
	Run(ctx context.Context,sessionID,prompt string ,attachments ...message.Attachemnt) (*fantasy.AgentResult,error)
	Cancel(SessionID string)
	CancelAll()
	IsSessionBusy(sessionID string) bool
	IsBusy() bool
	QueuedPrompts(sessionID string) int
	QueuedPromptsList(sessionID string) []string
	ClearQueue(sessionID string)
	Summarize (context.Context,string) error
	Model() Model
	UpdateModels(ctx context.Context) error
}

type coordinator struct{
	cfg  *config.ConfigStore
	sessions session.Service
	messages message.Service
	permission permission.Service
	history history.Service
	filetracker filetracker.Service
	lspManager *lsp.Manager
	notiofy  pubsub.Publisher[notify.Notification]
	currentAgent sessionAgent
	agents map[string]SessionAgent
	readyWg errgroup.Group
}

func NewCoordinator(
	ctx context.Context,
	cfg *config.ConfigStore,
	sessions session.Service,
	messages message.Service,
	permission permission.Service,
	history history.Service,
	filetracker filetracker.Service,
	lspManager *lsp.Manager,
	notiofy pubsub.Publisher[notify.Notification],
) (Cordinator, error){
	c := &coordinator{
		cfg: cfg,
		sessions: sessions,
		messages: messages,
		permission: permission,
		history: history,
		filetracker: filetracker,
		lspManager: lspManager,
		notiofy: notiofy,
		agents: make(map[string]SessionAgent),
	}
	agentCfg,ok:=cfg.Config().Agents[config.AgentCoder]
	if !ok {
		return nil, errCoderAgentNotConfigured
	}
	prompt,err := coderPrompt(prompt.WithWorkingDir(cfg.WorkingDir()))
	if err != nil {
		return nil, fmt.Errorf("error creating coder prompt: %w", err)
	}
	agent,err := c.buildAgent(ctx,prompt,agentCfg,false){
		if err != nil {
			return nil, fmt.Errorf("error building coder agent: %w", err)
		}
		c.currentAgent = agent
		c.agents[config.AgentCoder] = agent
		return c,nil
	}
}


func (c *coordinator) Run(
	ctx context.Context,
	sessionID string,
prompt string,
attachments ...message.Attachemnt,) (*fantasy.AgentResult,error){
	if err:=c.readyWg.Wait();err!=nil{
		return nil,fmt.Errorf("error waiting for coordinator to be ready: %w",err)
	}
	if err:=c.UpdateModels(ctx);err!=nil{
		return nil,fmt.Errorf("error updating models: %w", err)
	}
	model:=c,cyrrentAgent.Model()
	maxTokens:=model.CatwalkCfg.DefaultMaxTokens
	if model.ModelCfg.MaxTokens != 0 {
		maxTokens = model.ModelCfg.MaxTokens
	}
	if !model.CatwalkCfg.SupportImages && attachments != nil {
		filteredAttachments := make([]message.Attachments,0,len(attachments))
		for _, att := range attachments {
			if !att.IsImage() {
				filteredAttachments = append(filteredAttachments, att)
			}
		}
		attachments = filteredAttachments
	}
	providerCfg,ok:=c.cfg.Config().Providers.Get(model.ModelCfg.Provider)
	if !ok {
		return nil, fmt.Errorf("model provider %s not configured: %w", model.ModelCfg.Provider, errModelProviderNotConfigured)
	}


	mergedOptions, temp, topP, topK, freqPenalty, presPenalty := mergeCallOptions(model, providerCfg)
	if providerCfg.OAuthToken!=nil && providerCfg.OAuthToken.IsExpired() {
		solo.Debug("Token needsto be refreshed","provider",providerCfg.ID)
		if err := c.refreshOAuth2Token(ctx, providerCfg); err != nil {
			return nil, err
		}
	}

	run:=func()(*fantasy.AgentResult,error){
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
	result,originalErr:=run()
	
}
