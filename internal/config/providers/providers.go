// providers.go: top-level index of all first-party provider definitions.
//
// AllProviders returns every free + paid provider Ghost supports out of the box.
// Callers (config.Load) append these to the runtime provider list so users see
// them in /model, /provider, and ghost free without any extra config.
package providers

import "charm.land/catwalk/pkg/catwalk"

// AllFreeProviders returns every provider where at least one model is free.
func AllFreeProviders() []catwalk.Provider {
	return []catwalk.Provider{
		OpenCodeZenProvider,
		OIVSCodeProvider,
	}
}

// AllProviders returns free + paid providers.
func AllProviders() []catwalk.Provider {
	return AllFreeProviders() // current set is all free
}

// AllFreeModelIDs returns the (provider_id, model_id) tuple of every free model
// across all providers. Used by:
//   - ghost free picker
//   - dynamic_providers.go to mark models as free
//   - the TUI /model dialog to badge them
func AllFreeModelIDs() map[string]map[string]struct{} {
	out := map[string]map[string]struct{}{
		"opencode-zen": OpenCodeZenFreeModelIDs(),
		"nvidia-nim":   toSet(NvidiaNimFreeModelIDs),
	}
	return out
}

func toSet(ids []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		out[id] = struct{}{}
	}
	return out
}
