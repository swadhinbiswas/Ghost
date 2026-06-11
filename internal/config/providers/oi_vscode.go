// oi_vscode.go: OI VSCode Server gateway, pre-configured for MiniMax M2.7.
//
// Base URL: https://oi-vscode-server-985058387028.europe-west1.run.app
// Endpoint: /chat/completions (OpenAI-compatible)
// Auth:     none required (server accepts any bearer token; no key needed).
// Requires custom headers: customerId, userId, version.
package providers

import "charm.land/catwalk/pkg/catwalk"

// OIVSCodeProvider is the OI VSCode Server gateway providing MiniMax M2.7.
var OIVSCodeProvider = catwalk.Provider{
	ID:                  "oi-vscode",
	Name:                "OI VSCode Server (MiniMax)",
	APIEndpoint:         "https://oi-vscode-server-985058387028.europe-west1.run.app",
	APIKey:              "no-key-needed",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "openrouter/minimax-m2-thinking",
	DefaultSmallModelID: "openrouter/minimax-m2-thinking",
	DefaultHeaders: map[string]string{
		"customerId": "00000000-0000-0000-0000-000000000001",
		"userId":     "00000000-0000-0000-0000-000000000002",
		"version":    "1.1",
	},
	Models: []catwalk.Model{
		{
			ID:               "openrouter/minimax-m2-thinking",
			Name:             "MiniMax M2.7 (thinking)",
			ContextWindow:    200_000,
			DefaultMaxTokens: 8192,
			CanReason:        true,
		},
	},
}
