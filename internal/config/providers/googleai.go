// googleai.go: Google AI Studio (Gemini) provider.
//
// Base URL: https://generativelanguage.googleapis.com/v1beta
// Auth: Bearer $GEMINI_API_KEY (or x-goog-api-key header)
// Free tier: 60 requests/minute, generous daily quota.
package providers

import "charm.land/catwalk/pkg/catwalk"

var GoogleAIProvider = catwalk.Provider{
	ID:                  "googleai",
	Name:                "Google AI Studio (Gemini, free tier)",
	APIEndpoint:         "https://generativelanguage.googleapis.com/v1beta",
	APIKey:              "$GEMINI_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "gemini-2.5-pro",
	DefaultSmallModelID: "gemini-2.5-flash",
	Models: []catwalk.Model{
		{ID: "gemini-2.5-pro", Name: "Gemini 2.5 Pro (free)", ContextWindow: 2_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-2.5-flash", Name: "Gemini 2.5 Flash (free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-2.5-flash-lite", Name: "Gemini 2.5 Flash Lite (free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash (free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-2.0-flash-lite", Name: "Gemini 2.0 Flash Lite (free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro (deprecated)", ContextWindow: 2_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash (deprecated)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "gemini-embedding-001", Name: "Gemini Embedding 001", ContextWindow: 2_048, DefaultMaxTokens: 2_048},
	},
}
