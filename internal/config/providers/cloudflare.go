// cloudflare.go: Cloudflare Workers AI provider.
//
// Base URL: https://api.cloudflare.com/client/v4/accounts/{ACCOUNT_ID}/ai/v1
// Auth: Bearer $CLOUDFLARE_API_KEY
// Free tier: 10,000 neurons/day, plenty for a small dev workflow.
//
// Source: https://developers.cloudflare.com/workers-ai/
package providers

import "charm.land/catwalk/pkg/catwalk"

var CloudflareProvider = catwalk.Provider{
	ID:                  "cloudflare",
	Name:                "Cloudflare Workers AI (free 10K neurons/day)",
	APIEndpoint:         "https://api.cloudflare.com/client/v4/accounts/$CLOUDFLARE_ACCOUNT_ID/ai/v1",
	APIKey:              "$CLOUDFLARE_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "@cf/meta/llama-3.3-70b-instruct-fp8-fast",
	DefaultSmallModelID: "@cf/meta/llama-3.1-8b-instruct-fast",
	Models: []catwalk.Model{
		{ID: "@cf/meta/llama-3.3-70b-instruct-fp8-fast", Name: "Llama 3.3 70B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192},
		{ID: "@cf/meta/llama-3.1-8b-instruct-fast", Name: "Llama 3.1 8B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192},
		{ID: "@cf/meta/llama-3.1-70b-instruct", Name: "Llama 3.1 70B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192},
		{ID: "@cf/mistralai/mistral-7b-instruct-v0.1", Name: "Mistral 7B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192},
		{ID: "@cf/google/gemma-2-9b-it", Name: "Gemma 2 9B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192},
		{ID: "@cf/qwen/qwen2.5-coder-32b-instruct", Name: "Qwen 2.5 Coder 32B (Cloudflare)", ContextWindow: 24_000, DefaultMaxTokens: 8_192},
		{ID: "@cf/deepseek-ai/deepseek-r1-distill-qwen-32b", Name: "DeepSeek R1 Distill 32B (Cloudflare)", ContextWindow: 24_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "@cf/openai/gpt-oss-120b", Name: "GPT OSS 120B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "@cf/openai/gpt-oss-20b", Name: "GPT OSS 20B (Cloudflare, free)", ContextWindow: 24_000, DefaultMaxTokens: 8_192, CanReason: true},
	},
}
