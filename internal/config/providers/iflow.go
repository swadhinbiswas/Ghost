// iflow.go: iFlow provider.
//
// iFlow was a free unlimited provider but Alibaba discontinued the free OAuth
// tier in 2026. Kept here for compatibility with existing keys.
//
// Base URL: https://api.iflow.cn/v1
// Auth: Bearer $IFLOW_API_KEY
package providers

import "charm.land/catwalk/pkg/catwalk"

var IFlowProvider = catwalk.Provider{
	ID:                  "iflow",
	Name:                "iFlow",
	APIEndpoint:         "https://api.iflow.cn/v1",
	APIKey:              "$IFLOW_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "kimi-k2-thinking",
	DefaultSmallModelID: "qwen3-coder-plus",
	Models: []catwalk.Model{
		{ID: "kimi-k2-thinking", Name: "Kimi K2 Thinking (iFlow)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "kimi-k2", Name: "Kimi K2 (iFlow)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen3-coder-plus", Name: "Qwen 3 Coder Plus (iFlow)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "qwen3-max", Name: "Qwen 3 Max (iFlow)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "deepseek-v3", Name: "DeepSeek V3 (iFlow)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "deepseek-r1", Name: "DeepSeek R1 (iFlow)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "glm-4.6", Name: "GLM 4.6 (iFlow)", ContextWindow: 200_000, DefaultMaxTokens: 8_192},
	},
}
