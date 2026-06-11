// qwen.go: Alibaba Qwen (DashScope) OpenAI-compatible provider.
//
// Base URL: https://dashscope.aliyuncs.com/compatible-mode/v1
// Auth: Bearer $QWEN_API_KEY
// Free tier: 1M tokens for Qwen3-Coder-Plus, Qwen3-Max, Qwen-Plus.
//
// Source: https://help.aliyun.com/zh/model-studio/developer-reference/use-qwen-by-calling-api
package providers

import "charm.land/catwalk/pkg/catwalk"

var QwenProvider = catwalk.Provider{
	ID:                  "qwen",
	Name:                "Qwen (Alibaba DashScope)",
	APIEndpoint:         "https://dashscope.aliyuncs.com/compatible-mode/v1",
	APIKey:              "$QWEN_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "qwen3-coder-plus",
	DefaultSmallModelID: "qwen-turbo",
	Models: []catwalk.Model{
		{ID: "qwen3-coder-plus", Name: "Qwen 3 Coder Plus (free 1M)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen3-coder-plus-free", Name: "Qwen 3 Coder Plus Free", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen3-max", Name: "Qwen 3 Max (free trial)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "qwen3-max-preview", Name: "Qwen 3 Max Preview", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "qwen3-coder-480b-a35b-instruct", Name: "Qwen 3 Coder 480B", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen-plus", Name: "Qwen Plus (free 1M)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen-turbo", Name: "Qwen Turbo (free 1M)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen-long", Name: "Qwen Long (free 1M)", ContextWindow: 1_000_000, DefaultMaxTokens: 6_000},
		{ID: "qwq-32b", Name: "QwQ 32B (reasoning)", ContextWindow: 131_072, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "qwen3-next-80b-a3b-instruct", Name: "Qwen 3 Next 80B", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen3-next-80b-a3b-thinking", Name: "Qwen 3 Next 80B (thinking)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
	},
}
