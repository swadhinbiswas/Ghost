// nvidia_nim_free.go: free-tier NVIDIA NIM models on build.nvidia.com.
package providers

import "charm.land/catwalk/pkg/catwalk"

// NvidiaNimFreeModelIDs are the free-tier NVIDIA NIM model IDs.
var NvidiaNimFreeModelIDs = []string{
	"nvidia/nemotron-3-ultra-550b-a55b",
	"stepfun-ai/step-3.7-flash",
	"moonshotai/kimi-k2.6",
	"mistralai/mistral-medium-3.5-128b",
	"deepseek-ai/deepseek-v4-flash",
	"z-ai/glm-5.1",
	"minimaxai/minimax-m2.7",
	"google/gemma-4-31b-it",
	"nvidia/nemotron-3-super-120b-a12b",
	"qwen/qwen3.5-122b-a10b",
	"qwen/qwen3.5-397b-a17b",
}

// NvidiaNimFreeModels returns the catwalk.Model slice for the free NIM models.
func NvidiaNimFreeModels() []catwalk.Model {
	return []catwalk.Model{
		{ID: "nvidia/nemotron-3-ultra-550b-a55b", Name: "Nemotron 3 Ultra 550B (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "stepfun-ai/step-3.7-flash", Name: "Step 3.7 Flash (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "moonshotai/kimi-k2.6", Name: "Kimi K2.6 (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "mistralai/mistral-medium-3.5-128b", Name: "Mistral Medium 3.5 128B (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "deepseek-ai/deepseek-v4-flash", Name: "DeepSeek V4 Flash (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "z-ai/glm-5.1", Name: "GLM 5.1 (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "minimaxai/minimax-m2.7", Name: "MiniMax M2.7 (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "google/gemma-4-31b-it", Name: "Gemma 4 31B IT (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "nvidia/nemotron-3-super-120b-a12b", Name: "Nemotron 3 Super 120B (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CanReason: true, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "qwen/qwen3.5-122b-a10b", Name: "Qwen 3.5 122B A10B (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CostPer1MIn: 0, CostPer1MOut: 0},
		{ID: "qwen/qwen3.5-397b-a17b", Name: "Qwen 3.5 397B A17B (NIM, free)", ContextWindow: 128_000, DefaultMaxTokens: 16_384, CostPer1MIn: 0, CostPer1MOut: 0},
	}
}
