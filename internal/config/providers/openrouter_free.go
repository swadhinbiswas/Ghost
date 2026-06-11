// openrouter_free.go: OpenRouter provider pre-configured for free models.
//
// Base URL: https://openrouter.ai/api/v1
// Auth: Bearer $OPENROUTER_API_KEY
// Free models are tagged with :free suffix in OpenRouter.
package providers

import "charm.land/catwalk/pkg/catwalk"

var OpenRouterFreeProvider = catwalk.Provider{
	ID:                  "openrouter",
	Name:                "OpenRouter (free tier + paid)",
	APIEndpoint:         "https://openrouter.ai/api/v1",
	APIKey:              "$OPENROUTER_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "meta-llama/llama-3.3-70b-instruct:free",
	DefaultSmallModelID: "google/gemini-2.0-flash-exp:free",
	Models: []catwalk.Model{
		// === FREE TIER (`:free` suffix in OpenRouter) ===
		{ID: "meta-llama/llama-3.3-70b-instruct:free", Name: "Llama 3.3 70B (OpenRouter, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "meta-llama/llama-3.2-11b-vision-instruct:free", Name: "Llama 3.2 11B Vision (OpenRouter, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "google/gemini-2.0-flash-exp:free", Name: "Gemini 2.0 Flash Exp (OpenRouter, free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, SupportsImages: true},
		{ID: "mistralai/mistral-7b-instruct:free", Name: "Mistral 7B (OpenRouter, free)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "qwen/qwen-2.5-coder-32b-instruct:free", Name: "Qwen 2.5 Coder 32B (OpenRouter, free)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "nvidia/llama-3.1-nemotron-70b-instruct:free", Name: "Nemotron 70B (OpenRouter, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "deepseek/deepseek-chat:free", Name: "DeepSeek V3 (OpenRouter, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "openchat/openchat-7b:free", Name: "OpenChat 7B (OpenRouter, free)", ContextWindow: 8_192, DefaultMaxTokens: 8_192},
		{ID: "nousresearch/hermes-3-llama-3.1-405b:free", Name: "Hermes 3 405B (OpenRouter, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "google/gemma-2-9b-it:free", Name: "Gemma 2 9B IT (OpenRouter, free)", ContextWindow: 8_192, DefaultMaxTokens: 8_192},
	},
}
