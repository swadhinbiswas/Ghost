// cerebras.go: Cerebras Inference provider (free dev tier, ultra-fast).
//
// Base URL: https://api.cerebras.ai/v1
// Auth: Bearer $CEREBRAS_API_KEY
package providers

import "charm.land/catwalk/pkg/catwalk"

var CerebrasProvider = catwalk.Provider{
	ID:                  "cerebras",
	Name:                "Cerebras (free dev tier, ultra-fast)",
	APIEndpoint:         "https://api.cerebras.ai/v1",
	APIKey:              "$CEREBRAS_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "llama-3.3-70b",
	DefaultSmallModelID: "llama-3.1-8b",
	Models: []catwalk.Model{
		{ID: "llama-3.3-70b", Name: "Llama 3.3 70B (Cerebras, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "llama-3.1-70b", Name: "Llama 3.1 70B (Cerebras, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "llama-3.1-8b", Name: "Llama 3.1 8B (Cerebras, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "gpt-oss-120b", Name: "GPT OSS 120B (Cerebras, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "qwen-3-32b", Name: "Qwen 3 32B (Cerebras, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "deepseek-r1-distill-llama-70b", Name: "DeepSeek R1 Distill 70B (Cerebras, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
	},
}
