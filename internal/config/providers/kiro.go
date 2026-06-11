// kiro.go: Kiro AI provider (free Claude / GLM / MiniMax via OAuth).
//
// Kiro offers unlimited free usage of multiple models through OAuth login
// (GitHub, Google, AWS Builder ID, or AWS IAM Identity Center).
//
// Base URL: https://api.kiro.ai/v1 (OpenAI-compatible)
// Auth: Bearer $KIRO_API_KEY (OAuth-derived)
//
// Source: https://9router.com (which proxies Kiro as kr/ prefix)
package providers

import "charm.land/catwalk/pkg/catwalk"

var KiroProvider = catwalk.Provider{
	ID:                  "kiro",
	Name:                "Kiro AI (free, OAuth)",
	APIEndpoint:         "https://api.kiro.ai/v1",
	APIKey:              "$KIRO_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "claude-sonnet-4-5",
	DefaultSmallModelID: "claude-haiku-4-5",
	Models: []catwalk.Model{
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5 (Kiro, free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5 (Kiro, free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192},
		{ID: "glm-5", Name: "GLM-5 (Kiro, free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "MiniMax-M2.5", Name: "MiniMax M2.5 (Kiro, free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "qwen3-coder-next", Name: "Qwen3 Coder Next (Kiro, free)", ContextWindow: 262_144, DefaultMaxTokens: 8_192},
		{ID: "deepseek-3.2", Name: "DeepSeek 3.2 (Kiro, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
	},
}
