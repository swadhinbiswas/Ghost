// Package providers contains first-party provider definitions for Ghost.
//
// opencode_zen.go: the OpenCode Zen AI gateway. Source of truth:
//
//	https://opencode.ai/zen/v1/models
//	https://opencode.ai/docs/zen
//
// OpenCode Zen is a curated list of models provided by the OpenCode team.
// The endpoint is OpenAI-compatible. Free models (suffixed `-free` or named
// "Big Pickle") require no payment; paid models are charged pay-as-you-go.
package providers

import "charm.land/catwalk/pkg/catwalk"

// OpenCodeZenProvider is the OpenCode Zen gateway provider.
//
// Base URL:   https://opencode.ai/zen/v1
// Auth:       Authorization: Bearer $OPENCODE_API_KEY
// Free:       yes for the 5 free models (Big Pickle, DeepSeek V4 Flash Free,
//
//	MiMo-V2.5 Free, Qwen3.6 Plus Free, Nemotron 3 Super Free);
//	pay-as-you-go for the rest.
var OpenCodeZenProvider = catwalk.Provider{
	ID:                  "opencode-zen",
	Name:                "OpenCode Zen",
	APIEndpoint:         "https://opencode.ai/zen/v1",
	APIKey:              "no-key-needed",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "big-pickle",
	DefaultSmallModelID: "deepseek-v4-flash-free",
	Models: []catwalk.Model{
		// === FREE TIER (no payment required for these model IDs) ===
		{
			ID: "big-pickle", Name: "Big Pickle (stealth free)",
			ContextWindow: 200_000, DefaultMaxTokens: 8192,
		},
		{
			ID: "deepseek-v4-flash-free", Name: "DeepSeek V4 Flash Free",
			ContextWindow: 128_000, DefaultMaxTokens: 8192,
		},
		{
			ID: "mimo-v2.5-free", Name: "Xiaomi MiMo V2.5 Free",
			ContextWindow: 128_000, DefaultMaxTokens: 8192,
		},
		{
			ID: "qwen3.6-plus-free", Name: "Qwen 3.6 Plus Free",
			ContextWindow: 128_000, DefaultMaxTokens: 8192,
		},
		{
			ID: "nemotron-3-super-free", Name: "NVIDIA Nemotron 3 Super Free",
			ContextWindow: 1_000_000, DefaultMaxTokens: 8192, CanReason: true,
		},
	},
}

// FreeModelIDs returns the set of OpenCode Zen model IDs that are free.
func OpenCodeZenFreeModelIDs() map[string]struct{} {
	return map[string]struct{}{
		"big-pickle":             {},
		"deepseek-v4-flash-free": {},
		"mimo-v2.5-free":         {},
		"qwen3.6-plus-free":      {},
		"nemotron-3-super-free":  {},
	}
}
