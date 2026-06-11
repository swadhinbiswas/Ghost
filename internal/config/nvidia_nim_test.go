package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"charm.land/catwalk/pkg/catwalk"
)

func TestNvidiaNimDocToProvider(t *testing.T) {
	doc := nvidiaNimProviderDoc{
		ID:                  "nvidia-nim",
		Name:                "Nvidia NIM",
		APIEndpoint:         "https://integrate.api.nvidia.com/v1",
		APIKey:              "$NVIDIA_API_KEY",
		Type:                "openai-compat",
		DefaultLargeModelID: "nvidia/nemotron-3-ultra-550b-a55b",
		DefaultSmallModelID: "deepseek-ai/deepseek-v4-flash",
		Models: []nvidiaNimModelDoc{
			{ID: "nvidia/nemotron-3-ultra-550b-a55b", Name: "Nemotron 3 Ultra", ContextWindow: 128000, DefaultMaxTokens: 16384, CanReason: true},
			{ID: "deepseek-ai/deepseek-v4-flash", Name: "DeepSeek V4 Flash", ContextWindow: 128000, DefaultMaxTokens: 16384},
			{ID: "deepseek-ai/deepseek-v4-flash", Name: "dup should be dropped"}, // duplicate
		},
	}

	p, err := doc.toProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(p.ID) != "nvidia-nim" {
		t.Errorf("ID = %q, want nvidia-nim", p.ID)
	}
	if p.Type != catwalk.TypeOpenAICompat {
		t.Errorf("Type = %q, want OpenAI-compat", p.Type)
	}
	if len(p.Models) != 2 {
		t.Fatalf("Models = %d, want 2 (duplicate removed)", len(p.Models))
	}
	if p.DefaultLargeModelID != "nvidia/nemotron-3-ultra-550b-a55b" {
		t.Errorf("DefaultLargeModelID = %q", p.DefaultLargeModelID)
	}
}

func TestNvidiaNimDocDefaultsFallback(t *testing.T) {
	doc := nvidiaNimProviderDoc{
		ID: "nvidia-nim",
		Models: []nvidiaNimModelDoc{
			{ID: "only/model"},
		},
		DefaultLargeModelID: "does-not-exist",
		DefaultSmallModelID: "also-missing",
	}
	p, err := doc.toProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.DefaultLargeModelID != "only/model" || p.DefaultSmallModelID != "only/model" {
		t.Errorf("defaults not backfilled: large=%q small=%q", p.DefaultLargeModelID, p.DefaultSmallModelID)
	}
	// missing context/max tokens should be defaulted to sane values
	if p.Models[0].ContextWindow <= 0 || p.Models[0].DefaultMaxTokens <= 0 {
		t.Errorf("model metadata not defaulted: %+v", p.Models[0])
	}
	if p.APIEndpoint == "" || p.APIKey == "" {
		t.Errorf("endpoint/key not defaulted: %q %q", p.APIEndpoint, p.APIKey)
	}
}

func TestNvidiaNimDocEmptyModelsError(t *testing.T) {
	doc := nvidiaNimProviderDoc{ID: "nvidia-nim"}
	if _, err := doc.toProvider(); err == nil {
		t.Fatal("expected error for empty model list, got nil")
	}
}

func TestFetchNvidiaNimProvider(t *testing.T) {
	const body = `{
		"id": "nvidia-nim",
		"name": "Nvidia NIM",
		"api_endpoint": "https://integrate.api.nvidia.com/v1",
		"api_key": "$NVIDIA_API_KEY",
		"type": "openai-compat",
		"default_large_model_id": "nvidia/nemotron-3-ultra-550b-a55b",
		"default_small_model_id": "deepseek-ai/deepseek-v4-flash",
		"models": [
			{"id": "nvidia/nemotron-3-ultra-550b-a55b", "name": "Nemotron 3 Ultra 550B (NIM)", "context_window": 128000, "default_max_tokens": 16384, "can_reason": true},
			{"id": "deepseek-ai/deepseek-v4-flash", "name": "DeepSeek V4 Flash (NIM)", "context_window": 128000, "default_max_tokens": 16384, "can_reason": true}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	t.Setenv("GHOST_NVIDIA_NIM_MODELS_URL", srv.URL)

	p, err := fetchNvidiaNimProvider()
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(p.Models) != 2 {
		t.Errorf("Models = %d, want 2", len(p.Models))
	}
	if string(p.ID) != "nvidia-nim" {
		t.Errorf("ID = %q", p.ID)
	}
}

func TestFetchNvidiaNimProviderBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv("GHOST_NVIDIA_NIM_MODELS_URL", srv.URL)

	if _, err := fetchNvidiaNimProvider(); err == nil {
		t.Fatal("expected error on 404, got nil")
	}
}
