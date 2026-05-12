package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"charm.land/catwalk/pkg/catwalk"
)

// httpClient is reused across dynamic fetches.
var httpClient = &http.Client{Timeout: 15 * time.Second}

// openCodeModelsResponse matches the OpenCode /models endpoint.
type openCodeModelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

var openCodeLargeModelPreferences = []string{
	"big-pickle",
}

var openCodeSmallModelPreferences = []string{
	"deepseek-v4-flash-free",
	"minimax-m2.5-free",
	"ring-2.6-1t-free",
	"nemotron-3-super-free",
}

var openCodeDisplayNameOverrides = map[string]string{
	"big-pickle":             "Big Pickle",
	"deepseek-v4-flash-free": "DeepSeek V4 Flash Free",
	"minimax-m2.5-free":      "MiniMax M2.5 Free",
	"ring-2.6-1t-free":       "Ring 2.6 1T Free",
	"nemotron-3-super-free":  "Nemotron 3 Super Free",
}

var openCodeFreeModelIDs = map[string]struct{}{
	"big-pickle":             {},
	"deepseek-v4-flash-free": {},
	"minimax-m2.5-free":      {},
	"ring-2.6-1t-free":       {},
	"nemotron-3-super-free":  {},
}

// fetchOpenCodeModels returns the models exposed by the OpenCode Zen model list.
func fetchOpenCodeModels() ([]catwalk.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://opencode.ai/zen/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("opencode: create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("opencode: fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("opencode: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var parsed openCodeModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("opencode: decode response: %w", err)
	}

	var (
		models []catwalk.Model
		seen   = make(map[string]struct{})
	)
	for _, m := range parsed.Data {
		if m.ID == "" {
			continue
		}
		if _, ok := openCodeFreeModelIDs[m.ID]; !ok {
			continue
		}
		if _, ok := seen[m.ID]; ok {
			continue
		}
		seen[m.ID] = struct{}{}
		models = append(models, catwalk.Model{
			ID:               m.ID,
			Name:             displayNameFromOpenCodeID(m.ID),
			ContextWindow:    128000,
			DefaultMaxTokens: 8192,
		})
	}

	sortOpenCodeModels(models)
	return models, nil
}

func sortOpenCodeModels(models []catwalk.Model) {
	priority := func(id string) int {
		for i, preferred := range append(openCodeLargeModelPreferences, openCodeSmallModelPreferences...) {
			if id == preferred {
				return i
			}
		}
		if strings.Contains(id, "free") {
			return len(openCodeLargeModelPreferences) + len(openCodeSmallModelPreferences)
		}
		return len(openCodeLargeModelPreferences) + len(openCodeSmallModelPreferences) + 1
	}

	sort.SliceStable(models, func(i, j int) bool {
		pi := priority(models[i].ID)
		pj := priority(models[j].ID)
		if pi != pj {
			return pi < pj
		}
		return models[i].ID < models[j].ID
	})
}

func displayNameFromOpenCodeID(id string) string {
	if name, ok := openCodeDisplayNameOverrides[id]; ok {
		return name
	}
	return displayNameFromID(id)
}

func pickOpenCodeModelID(models []catwalk.Model, preferredIDs []string) string {
	if len(models) == 0 {
		return ""
	}

	lookup := make(map[string]struct{}, len(models))
	for _, model := range models {
		lookup[model.ID] = struct{}{}
	}
	for _, preferredID := range preferredIDs {
		if _, ok := lookup[preferredID]; ok {
			return preferredID
		}
	}

	return models[0].ID
}

func pickOpenCodeSmallModelID(models []catwalk.Model, largeModelID string) string {
	if len(models) == 0 {
		return ""
	}

	lookup := make(map[string]struct{}, len(models))
	for _, model := range models {
		lookup[model.ID] = struct{}{}
	}
	for _, preferredID := range openCodeSmallModelPreferences {
		if preferredID == largeModelID {
			continue
		}
		if _, ok := lookup[preferredID]; ok {
			return preferredID
		}
	}

	for _, model := range models {
		if model.ID != largeModelID {
			return model.ID
		}
	}

	return largeModelID
}

// openRouterModelsResponse matches the OpenRouter /api/v1/models endpoint.
type openRouterModelsResponse struct {
	Data []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ContextLength int64  `json:"context_length"`
		TopProvider   struct {
			MaxCompletionTokens int64 `json:"max_completion_tokens"`
		} `json:"top_provider"`
		Pricing struct {
			Prompt string `json:"prompt"`
		} `json:"pricing"`
	} `json:"data"`
}

// fetchOpenRouterFreeModels returns only models whose IDs end with ":free".
// It also accepts a set of OpenCode normalized model names so duplicates can
// be filtered out (OpenCode free models have no rate limits, so we prefer them).
func fetchOpenRouterFreeModels(ocNormalized map[string]struct{}) ([]catwalk.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openrouter.ai/api/v1/models", nil)
	if err != nil {
		return nil, fmt.Errorf("openrouter: create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter: fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var parsed openRouterModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("openrouter: decode response: %w", err)
	}

	var models []catwalk.Model
	for _, m := range parsed.Data {
		if !strings.HasSuffix(m.ID, ":free") {
			continue
		}

		// Skip OpenRouter models that are also available for free on OpenCode
		// (OpenCode doesn't have the strict rate limits that OpenRouter free tier has).
		norm := normalizeModelID(m.ID)
		if _, isDup := ocNormalized[norm]; isDup {
			continue
		}

		name := m.Name
		if name == "" {
			name = displayNameFromID(m.ID)
		}
		ctxWindow := m.ContextLength
		if ctxWindow == 0 {
			ctxWindow = 128000
		}
		maxTok := m.TopProvider.MaxCompletionTokens
		if maxTok == 0 {
			maxTok = 8192
		}
		models = append(models, catwalk.Model{
			ID:               m.ID,
			Name:             name,
			ContextWindow:    ctxWindow,
			DefaultMaxTokens: maxTok,
		})
	}
	return models, nil
}

// sizeSuffixes matches common model size suffixes like "-8b", "-70b-instruct", "-120b-a12b".
var sizeSuffixes = regexp.MustCompile(`-\d+[Bb](?:-\w+)*$`)

// normalizeModelID strips provider prefixes, free suffixes, and size details so
// models can be compared across providers.
//
// Examples:
//
//	"hy3-preview-free"           -> "hy3 preview"
//	"tencent/hy3-preview:free"   -> "hy3 preview"
//	"minimax-m2.5-free"          -> "minimax m2.5"
//	"minimax/minimax-m2.5:free"  -> "minimax m2.5"
func normalizeModelID(id string) string {
	// Strip provider prefix (e.g. "tencent/", "nvidia/").
	if idx := strings.LastIndex(id, "/"); idx != -1 {
		id = id[idx+1:]
	}

	// Strip free suffixes.
	id = strings.TrimSuffix(id, ":free")
	id = strings.TrimSuffix(id, "-free")

	// Strip size suffixes like "-120b-a12b", "-8b-instruct".
	id = sizeSuffixes.ReplaceAllString(id, "")

	// Normalize separators to spaces.
	id = strings.ReplaceAll(id, "-", " ")
	id = strings.ReplaceAll(id, "_", " ")
	id = strings.ToLower(strings.TrimSpace(id))

	return id
}

// displayNameFromID converts a model ID like "minimax-m2.5-free" into "MiniMax M2.5 Free".
func displayNameFromID(id string) string {
	// Strip known suffixes.
	id = strings.TrimSuffix(id, ":free")
	id = strings.TrimSuffix(id, "-free")

	// Replace common separators with spaces.
	s := strings.ReplaceAll(id, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "/", " ")

	// Title-case each word.
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// fetchDynamicProviders returns providers whose model lists are fetched live from APIs.
// Errors are logged and the provider is omitted if the fetch fails.
func fetchDynamicProviders() []catwalk.Provider {
	var out []catwalk.Provider

	// 1. OpenCode Zen – fetch first so we can use its model list to deduplicate OpenRouter.
	ocModels, _ := fetchOpenCodeModels()
	if len(ocModels) > 0 {
		largeModelID := pickOpenCodeModelID(ocModels, openCodeLargeModelPreferences)
		smallModelID := pickOpenCodeSmallModelID(ocModels, largeModelID)
		out = append(out, catwalk.Provider{
			ID:                  "opencode",
			Name:                "OpenCode Zen",
			APIEndpoint:         "https://opencode.ai/zen/v1",
			Type:                catwalk.TypeOpenAICompat,
			DefaultLargeModelID: largeModelID,
			DefaultSmallModelID: smallModelID,
			Models:              ocModels,
		})
	}

	// Build a lookup set of normalized OpenCode model names.
	ocNormalized := make(map[string]struct{}, len(ocModels))
	for _, m := range ocModels {
		ocNormalized[normalizeModelID(m.ID)] = struct{}{}
	}

	// 2. OpenRouter – free models only, minus duplicates already offered by OpenCode.
	if orModels, err := fetchOpenRouterFreeModels(ocNormalized); err == nil && len(orModels) > 0 {
		out = append(out, catwalk.Provider{
			ID:                  "openrouter",
			Name:                "OpenRouter",
			APIEndpoint:         "https://openrouter.ai/api/v1",
			Type:                catwalk.TypeOpenAI,
			DefaultLargeModelID: orModels[0].ID,
			DefaultSmallModelID: orModels[0].ID,
			Models:              orModels,
		})
	}

	return out
}
