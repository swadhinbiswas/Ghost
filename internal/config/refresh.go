package config

import (
	"log/slog"
	"sync/atomic"
)

// dynamicRefreshRunning guards against concurrent Refresh_Cycles. Only one
// background refresh may run at a time; additional triggers are no-ops while a
// cycle is in flight.
var dynamicRefreshRunning atomic.Bool

// RefreshDynamicProvidersInBackground starts a single, non-blocking refresh of
// the dynamically-fetched provider model lists (OpenCode Zen + Nvidia NIM).
//
// The refresh runs on a background goroutine so startup is never blocked: the
// application keeps serving the model list it loaded from cache/static while
// the fresh list is fetched and written to the local cache for the next run.
//
// Behavior matches the opencode-zen-model-refresh spec:
//   - No-op when auto-update is disabled.
//   - No-op when a refresh is already running (concurrency guard).
//   - On any fetch failure, or a result with zero providers, the existing cache
//     is left unchanged and a diagnostic is logged. Never panics or blocks.
func RefreshDynamicProvidersInBackground(autoUpdateDisabled bool) {
	if autoUpdateDisabled {
		slog.Debug("Provider auto-update disabled; skipping background model refresh")
		return
	}
	if !dynamicRefreshRunning.CompareAndSwap(false, true) {
		slog.Debug("Model refresh already running; skipping duplicate trigger")
		return
	}

	go func() {
		defer dynamicRefreshRunning.Store(false)
		refreshDynamicProvidersOnce()
	}()
}

// refreshDynamicProvidersOnce performs one Refresh_Cycle synchronously. It is
// the unit of work executed by the background goroutine and is also directly
// testable.
func refreshDynamicProvidersOnce() {
	fresh := fetchFreshDynamicProviders()
	if len(fresh) == 0 {
		// All remote sources failed or returned nothing usable — keep the
		// existing cache contents unchanged.
		slog.Warn("Model refresh produced no providers; keeping existing cache")
		return
	}

	if err := saveDynamicProvidersToCache(fresh); err != nil {
		slog.Warn("Failed to persist refreshed model cache", "error", err)
		return
	}

	slog.Info("Refreshed dynamic provider model cache", "providers", len(fresh))
}
