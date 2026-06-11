// groq.go: Groq provider (free dev tier).
//
// Base URL: https://api.groq.com/openai/v1
// Auth: Bearer $GROQ_API_KEY
// Free tier: limited dev rate; very fast LPU inference.
package providers

import "charm.land/catwalk/pkg/catwalk"

var GroqProvider = catwalk.Provider{
	ID:                  "groq",
	Name:                "Groq (free dev tier, ultra-fast LPU)",
	APIEndpoint:         "https://api.groq.com/openai/v1",
	APIKey:              "$GROQ_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "llama-3.3-70b-versatile",
	DefaultSmallModelID: "llama-3.1-8b-instant",
	Models: []catwalk.Model{
		{ID: "llama-3.3-70b-versatile", Name: "Llama 3.3 70B Versatile (Groq, free)", ContextWindow: 128_000, DefaultMaxTokens: 32_768},
		{ID: "llama-3.1-70b-versatile", Name: "Llama 3.1 70B Versatile (Groq, free)", ContextWindow: 128_000, DefaultMaxTokens: 32_768},
		{ID: "llama-3.1-8b-instant", Name: "Llama 3.1 8B Instant (Groq, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "llama-3-8b-8192", Name: "Llama 3 8B (Groq, free)", ContextWindow: 8_192, DefaultMaxTokens: 8_192},
		{ID: "llama-3-70b-8192", Name: "Llama 3 70B (Groq, free)", ContextWindow: 8_192, DefaultMaxTokens: 8_192},
		{ID: "mixtral-8x7b-32768", Name: "Mixtral 8x7B (Groq, free)", ContextWindow: 32_768, DefaultMaxTokens: 32_768},
		{ID: "gemma2-9b-it", Name: "Gemma 2 9B IT (Groq, free)", ContextWindow: 8_192, DefaultMaxTokens: 8_192},
		{ID: "deepseek-r1-distill-llama-70b", Name: "DeepSeek R1 Distill 70B (Groq, free)", ContextWindow: 128_000, DefaultMaxTokens: 32_768, CanReason: true},
		{ID: "qwen-qwq-32b", Name: "QwQ 32B (Groq, free, reasoning)", ContextWindow: 128_000, DefaultMaxTokens: 32_768, CanReason: true},
	},
}
