// github_copilot_free.go: GitHub Copilot free models (subset of gh/ in 9Router).
//
// This is the standalone GitHub Copilot provider, not via 9Router.
// Base URL: https://api.githubcopilot.com
// Auth: Bearer $GITHUB_COPILOT_TOKEN (OAuth)
package providers

import "charm.land/catwalk/pkg/catwalk"

var GitHubCopilotFreeProvider = catwalk.Provider{
	ID:                  "github-copilot",
	Name:                "GitHub Copilot (free quota, OAuth)",
	APIEndpoint:         "https://api.githubcopilot.com",
	APIKey:              "$GITHUB_COPILOT_TOKEN",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "gpt-4o",
	DefaultSmallModelID: "gpt-4o-mini",
	Models: []catwalk.Model{
		{ID: "gpt-4o", Name: "GPT 4o (Copilot, free quota)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, SupportsImages: true},
		{ID: "gpt-4o-mini", Name: "GPT 4o mini (Copilot, free quota)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, SupportsImages: true},
		{ID: "o1-preview", Name: "o1 preview (Copilot, paid)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true},
		{ID: "o1-mini", Name: "o1 mini (Copilot, free quota)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true},
		{ID: "claude-3.5-sonnet", Name: "Claude 3.5 Sonnet (Copilot, free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192},
		{ID: "claude-3.7-sonnet", Name: "Claude 3.7 Sonnet (Copilot, paid)", ContextWindow: 200_000, DefaultMaxTokens: 16_384},
		{ID: "claude-sonnet-4", Name: "Claude Sonnet 4 (Copilot, paid)", ContextWindow: 200_000, DefaultMaxTokens: 16_384, CanReason: true},
		{ID: "gemini-2.0-flash", Name: "Gemini 2.0 Flash (Copilot, free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192},
		{ID: "grok-code-fast-1", Name: "Grok Code Fast 1 (Copilot, free)", ContextWindow: 131_072, DefaultMaxTokens: 8_192},
	},
}
