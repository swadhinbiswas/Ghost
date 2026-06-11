package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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
	"mimo-v2.5-free",
	"qwen3.6-plus-free",
	"north-mini-code-free",
}

var openCodeDisplayNameOverrides = map[string]string{
	"big-pickle":             "Big Pickle",
	"deepseek-v4-flash-free": "DeepSeek V4 Flash Free",
	"mimo-v2.5-free":         "Xiaomi MiMo V2.5 Free",
	"qwen3.6-plus-free":      "Qwen 3.6 Plus Free",
	"minimax-m3-free":        "MiniMax M3 Free",
	"nemotron-3-ultra-free":  "Nemotron 3 Ultra Free",
	"north-mini-code-free":   "North Mini Code Free",
}

// openCodeFreeModelIDs is a fallback hint set used only for offline/static
// rendering. The authoritative free-model detection is isOpenCodeFreeModel,
// which is applied to the live endpoint response.
var openCodeFreeModelIDs = map[string]struct{}{
	"big-pickle":             {},
	"deepseek-v4-flash-free": {},
	"mimo-v2.5-free":         {},
	"qwen3.6-plus-free":      {},
	"minimax-m3-free":        {},
	"nemotron-3-ultra-free":  {},
	"north-mini-code-free":   {},
}

// isOpenCodeFreeModel reports whether an OpenCode Zen model ID is free.
// Free models are suffixed "-free"; "big-pickle" is a special stealth free model.
func isOpenCodeFreeModel(id string) bool {
	return id == "big-pickle" || strings.HasSuffix(id, "-free")
}

// openCodeModelsURL is the OpenCode Zen model list endpoint.
const openCodeModelsURL = "https://opencode.ai/zen/v1/models"

// openCodeModelsEndpoint returns the configured OpenCode Zen models URL,
// honoring the GHOST_OPENCODE_MODELS_URL override (used for testing/forks).
func openCodeModelsEndpoint() string {
	if v := strings.TrimSpace(os.Getenv("GHOST_OPENCODE_MODELS_URL")); v != "" {
		return v
	}
	return openCodeModelsURL
}

// fetchOpenCodeModels returns the models exposed by the OpenCode Zen model list.
func fetchOpenCodeModels() ([]catwalk.Model, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, openCodeModelsEndpoint(), nil)
	if err != nil {
		return nil, fmt.Errorf("opencode: create request: %w", err)
	}
	req.Header.Set("User-Agent", "OpenCode/1.0.0")

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
		if !isOpenCodeFreeModel(m.ID) {
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
// fetchDynamicProviders returns providers whose model lists are cached locally.
func fetchDynamicProviders() []catwalk.Provider {
	// Try loading from cache first to avoid slow network lookups on startup
	if cached, err := loadDynamicProvidersFromCache(); err == nil && len(cached) > 0 {
		return cached
	}

	// Fallback to static embedded/local providers if cache is empty/invalid
	return getStaticDynamicProvidersFallback()
}

// fetchFreshDynamicProviders fetches dynamic providers remotely. Each source is
// best-effort: a failure is logged and that provider is simply omitted from the
// fresh set (callers fall back to cache or embedded defaults).
func fetchFreshDynamicProviders() []catwalk.Provider {
	var out []catwalk.Provider

	// 1. OpenCode Zen – fetch first so we can use its model list to deduplicate OpenRouter.
	ocModels, err := fetchOpenCodeModels()
	if err != nil {
		slog.Warn("Failed to fetch OpenCode Zen models, will use cache/fallback", "error", err)
	}
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

	// 2. Nvidia NIM – fetch the live model catalog from the remote JSON source.
	if nim, err := fetchNvidiaNimProvider(); err != nil {
		slog.Warn("Failed to fetch Nvidia NIM models, will use cache/fallback", "error", err)
	} else {
		out = append(out, nim)
	}

	return out
}

func saveDynamicProvidersToCache(providers []catwalk.Provider) error {
	path := cachePathFor("dynamic_providers")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(providers)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func loadDynamicProvidersFromCache() ([]catwalk.Provider, error) {
	path := cachePathFor("dynamic_providers")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var providers []catwalk.Provider
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

func getStaticDynamicProvidersFallback() []catwalk.Provider {
	var out []catwalk.Provider

	// 1. OpenCode Zen static fallback
	out = append(out, catwalk.Provider{
		ID:                  "opencode",
		Name:                "OpenCode Zen",
		APIEndpoint:         "https://opencode.ai/zen/v1",
		APIKey:              "no-key-needed",
		Type:                catwalk.TypeOpenAICompat,
		DefaultLargeModelID: "big-pickle",
		DefaultSmallModelID: "deepseek-v4-flash-free",
		Models:              getStaticOpenCodeModels(),
	})

	return out
}

func getStaticOpenCodeModels() []catwalk.Model {
	return []catwalk.Model{
		{ID: "big-pickle", Name: "Big Pickle (stealth free)", ContextWindow: 200_000, DefaultMaxTokens: 8192},
		{ID: "deepseek-v4-flash-free", Name: "DeepSeek V4 Flash Free", ContextWindow: 128_000, DefaultMaxTokens: 8192},
		{ID: "mimo-v2.5-free", Name: "Xiaomi MiMo V2.5 Free", ContextWindow: 128_000, DefaultMaxTokens: 8192},
		{ID: "qwen3.6-plus-free", Name: "Qwen 3.6 Plus Free", ContextWindow: 128_000, DefaultMaxTokens: 8192},
		{ID: "minimax-m3-free", Name: "MiniMax M3 Free", ContextWindow: 128_000, DefaultMaxTokens: 8192, CanReason: true},
		{ID: "nemotron-3-ultra-free", Name: "Nemotron 3 Ultra Free", ContextWindow: 1_000_000, DefaultMaxTokens: 8192, CanReason: true},
		{ID: "north-mini-code-free", Name: "North Mini Code Free", ContextWindow: 128_000, DefaultMaxTokens: 8192},
	}
}
