package config

import "charm.land/catwalk/pkg/catwalk"

// getLocalProviders returns manually-configured providers used as the embedded
// fallback when the remote model catalog cannot be fetched.
//
// Nvidia NIM is the source-of-truth-by-remote provider: its authoritative model
// list lives in nvidia_nim_models.json (served from GitHub and fetched via
// fetchNvidiaNimProvider). This embedded copy mirrors that file so Ghost works
// offline and on first run before the cache is populated.
func getLocalProviders() []catwalk.Provider {
	return []catwalk.Provider{
		{
			ID:                  "nvidia-nim",
			Name:                "Nvidia NIM",
			APIEndpoint:         "https://integrate.api.nvidia.com/v1",
			APIKey:              "$NVIDIA_API_KEY",
			Type:                catwalk.TypeOpenAICompat,
			DefaultLargeModelID: "nvidia/nemotron-3-ultra-550b-a55b",
			DefaultSmallModelID: "deepseek-ai/deepseek-v4-flash",
			Models: []catwalk.Model{
				{ID: "nvidia/nemotron-3-ultra-550b-a55b", Name: "Nemotron 3 Ultra 550B (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384, CanReason: true},
				{ID: "deepseek-ai/deepseek-v4-pro", Name: "DeepSeek V4 Pro (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384, CanReason: true},
				{ID: "deepseek-ai/deepseek-v4-flash", Name: "DeepSeek V4 Flash (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384, CanReason: true},
				{ID: "z-ai/glm-5.1", Name: "GLM 5.1 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384, CanReason: true},
				{ID: "moonshotai/kimi-k2.6", Name: "Kimi K2.6 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384},
				{ID: "qwen/qwen3.5-397b-a17b", Name: "Qwen 3.5 397B A17B (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384},
				{ID: "qwen/qwen3.5-122b-a10b", Name: "Qwen 3.5 122B A10B (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384},
				{ID: "mistralai/mistral-medium-3.5-128b", Name: "Mistral Medium 3.5 128B (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384, CanReason: true},
				{ID: "stepfun-ai/step-3.7-flash", Name: "Step 3.7 Flash (NIM)", ContextWindow: 128000, DefaultMaxTokens: 16384},
				{ID: "minimaxai/minimax-m2.7", Name: "MiniMax M2.7 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "google/diffusiongemma-26b-a4b-it", Name: "DiffusionGemma 26B A4B IT (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096, CanReason: true},
			},
		},
	}
}
