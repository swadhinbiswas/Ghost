package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"charm.land/catwalk/pkg/catwalk"
)

// nvidiaNimModelsURL is the canonical, remotely-updatable source of truth for
// the Nvidia NIM model catalog. It is a raw GitHub link so the list can be
// updated without shipping a new binary.
//
// Override at runtime with GHOST_NVIDIA_NIM_MODELS_URL (handy for testing or
// pointing at a fork/branch).
const nvidiaNimModelsURL = "https://raw.githubusercontent.com/swadhinbiswas/Ghost/main/nvidia_nim_models.json"

// nvidiaNimProviderDoc is the on-disk / remote JSON schema for the Nvidia NIM
// provider definition. Field names are stable and decoupled from catwalk's
// internal struct tags so the file format never breaks on a dependency bump.
type nvidiaNimProviderDoc struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	APIEndpoint         string              `json:"api_endpoint"`
	APIKey              string              `json:"api_key"`
	Type                string              `json:"type"`
	DefaultLargeModelID string              `json:"default_large_model_id"`
	DefaultSmallModelID string              `json:"default_small_model_id"`
	UpdatedAt           string              `json:"updated_at,omitempty"`
	Models              []nvidiaNimModelDoc `json:"models"`
}

type nvidiaNimModelDoc struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ContextWindow    int64  `json:"context_window"`
	DefaultMaxTokens int64  `json:"default_max_tokens"`
	CanReason        bool   `json:"can_reason,omitempty"`
}

// nvidiaNimModelsEndpoint returns the configured NIM models URL, honoring the
// GHOST_NVIDIA_NIM_MODELS_URL override.
func nvidiaNimModelsEndpoint() string {
	if v := strings.TrimSpace(os.Getenv("GHOST_NVIDIA_NIM_MODELS_URL")); v != "" {
		return v
	}
	return nvidiaNimModelsURL
}

// fetchNvidiaNimProvider retrieves the Nvidia NIM provider definition from the
// remote JSON source and maps it to a catwalk.Provider. On any failure the
// caller should fall back to the embedded static definition.
func fetchNvidiaNimProvider() (catwalk.Provider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := nvidiaNimModelsEndpoint()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return catwalk.Provider{}, fmt.Errorf("nvidia-nim: create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Ghost-CLI")

	resp, err := httpClient.Do(req)
	if err != nil {
		return catwalk.Provider{}, fmt.Errorf("nvidia-nim: fetch models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return catwalk.Provider{}, fmt.Errorf("nvidia-nim: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var doc nvidiaNimProviderDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return catwalk.Provider{}, fmt.Errorf("nvidia-nim: decode response: %w", err)
	}

	provider, err := doc.toProvider()
	if err != nil {
		return catwalk.Provider{}, fmt.Errorf("nvidia-nim: %w", err)
	}
	return provider, nil
}

// toProvider validates the document and converts it to a catwalk.Provider.
func (d nvidiaNimProviderDoc) toProvider() (catwalk.Provider, error) {
	if d.ID == "" {
		return catwalk.Provider{}, fmt.Errorf("missing provider id")
	}

	models := make([]catwalk.Model, 0, len(d.Models))
	seen := make(map[string]struct{}, len(d.Models))
	for _, m := range d.Models {
		if m.ID == "" {
			continue
		}
		if _, dup := seen[m.ID]; dup {
			continue
		}
		seen[m.ID] = struct{}{}

		name := m.Name
		if name == "" {
			name = m.ID
		}
		ctx := m.ContextWindow
		if ctx <= 0 {
			ctx = 128_000
		}
		maxTok := m.DefaultMaxTokens
		if maxTok <= 0 {
			maxTok = 8192
		}
		models = append(models, catwalk.Model{
			ID:               m.ID,
			Name:             name,
			ContextWindow:    ctx,
			DefaultMaxTokens: maxTok,
			CanReason:        m.CanReason,
		})
	}

	if len(models) == 0 {
		return catwalk.Provider{}, fmt.Errorf("no valid models in document")
	}

	endpoint := d.APIEndpoint
	if endpoint == "" {
		endpoint = "https://integrate.api.nvidia.com/v1"
	}
	apiKey := d.APIKey
	if apiKey == "" {
		apiKey = "$NVIDIA_API_KEY"
	}

	// Pick defaults, falling back to the first model when unset/invalid.
	large := d.DefaultLargeModelID
	if _, ok := seen[large]; !ok {
		large = models[0].ID
	}
	small := d.DefaultSmallModelID
	if _, ok := seen[small]; !ok {
		small = large
	}

	return catwalk.Provider{
		ID:                  catwalk.InferenceProvider(d.ID),
		Name:                cmpOr(d.Name, "Nvidia NIM"),
		APIEndpoint:         endpoint,
		APIKey:              apiKey,
		Type:                nimProviderType(d.Type),
		DefaultLargeModelID: large,
		DefaultSmallModelID: small,
		Models:              models,
	}, nil
}

// nimProviderType maps the JSON type string to a catwalk provider type,
// defaulting to OpenAI-compatible.
func nimProviderType(t string) catwalk.Type {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "", "openai-compat", "openai", "openai_compatible", "openaicompat":
		return catwalk.TypeOpenAICompat
	default:
		// Unknown types still default to OpenAI-compatible — NIM speaks it.
		return catwalk.TypeOpenAICompat
	}
}

// cmpOr returns a if non-empty, otherwise b.
func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
