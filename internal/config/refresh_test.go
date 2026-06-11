package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const testOpenCodeBody = `{
	"object": "list",
	"data": [
		{"id": "big-pickle", "object": "model"},
		{"id": "deepseek-v4-flash-free", "object": "model"},
		{"id": "some-paid-model", "object": "model"}
	]
}`

const testNimBody = `{
	"id": "nvidia-nim",
	"name": "Nvidia NIM",
	"api_endpoint": "https://integrate.api.nvidia.com/v1",
	"api_key": "$NVIDIA_API_KEY",
	"type": "openai-compat",
	"default_large_model_id": "nvidia/nemotron-3-ultra-550b-a55b",
	"default_small_model_id": "deepseek-ai/deepseek-v4-flash",
	"models": [
		{"id": "nvidia/nemotron-3-ultra-550b-a55b", "name": "Nemotron 3 Ultra", "context_window": 128000, "default_max_tokens": 16384, "can_reason": true}
	]
}`

// startStub serves the given body for every request.
func startStub(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestRefreshDynamicProvidersOnceWritesCache(t *testing.T) {
	// Point all remote sources at local stubs and the cache at a temp dir.
	t.Setenv("GHOST_OPENCODE_MODELS_URL", startStub(t, testOpenCodeBody))
	t.Setenv("GHOST_NVIDIA_NIM_MODELS_URL", startStub(t, testNimBody))
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	refreshDynamicProvidersOnce()

	cached, err := loadDynamicProvidersFromCache()
	if err != nil {
		t.Fatalf("cache not written: %v", err)
	}

	ids := map[string]bool{}
	for _, p := range cached {
		ids[string(p.ID)] = true
	}
	if !ids["opencode"] {
		t.Errorf("expected opencode provider in cache, got %v", ids)
	}
	if !ids["nvidia-nim"] {
		t.Errorf("expected nvidia-nim provider in cache, got %v", ids)
	}

	// OpenCode must contain only the two free models, not the paid one.
	for _, p := range cached {
		if string(p.ID) != "opencode" {
			continue
		}
		for _, m := range p.Models {
			if m.ID == "some-paid-model" {
				t.Errorf("paid model leaked into refreshed cache")
			}
		}
	}
}

func TestRefreshDynamicProvidersOnceFailureKeepsCache(t *testing.T) {
	// Both sources fail (404) and the cache dir is empty → nothing written.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	t.Setenv("GHOST_OPENCODE_MODELS_URL", srv.URL)
	t.Setenv("GHOST_NVIDIA_NIM_MODELS_URL", srv.URL)
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	refreshDynamicProvidersOnce()

	if _, err := loadDynamicProvidersFromCache(); err == nil {
		t.Fatal("expected no cache file to be written on total failure")
	}
}

func TestRefreshBackgroundConcurrencyGuard(t *testing.T) {
	// Simulate a refresh already in flight; a new trigger must be a no-op and
	// must not reset the running flag.
	if !dynamicRefreshRunning.CompareAndSwap(false, true) {
		t.Fatal("guard should start in the not-running state")
	}
	defer dynamicRefreshRunning.Store(false)

	RefreshDynamicProvidersInBackground(false) // should early-return

	if !dynamicRefreshRunning.Load() {
		t.Error("concurrency guard was unexpectedly cleared by a duplicate trigger")
	}
}

func TestRefreshBackgroundDisabledIsNoop(t *testing.T) {
	if dynamicRefreshRunning.Load() {
		t.Fatal("precondition: refresh should not be running")
	}
	RefreshDynamicProvidersInBackground(true) // disabled → no-op
	if dynamicRefreshRunning.Load() {
		t.Error("disabled refresh must not start a cycle")
	}
}
