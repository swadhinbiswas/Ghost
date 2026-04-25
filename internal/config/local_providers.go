package config

import "charm.land/catwalk/pkg/catwalk"

// getLocalProviders returns manually-configured providers that are NOT available
// in the Catwalk registry and CANNOT be fetched from a public API without auth.
// Currently only Nvidia NIM falls into this category.
func getLocalProviders() []catwalk.Provider {
	return []catwalk.Provider{
		{
			ID:                  "nvidia-nim",
			Name:                "Nvidia NIM",
			APIEndpoint:         "https://integrate.api.nvidia.com/v1",
			APIKey:              "$NVIDIA_API_KEY",
			Type:                catwalk.TypeOpenAICompat,
			DefaultLargeModelID: "z-ai/glm4.7",
			DefaultSmallModelID: "microsoft/phi-3-mini-128k-instruct",
			Models: []catwalk.Model{
				{ID: "deepseek-ai/deepseek-v4-flash", Name: "DeepSeek V4 Flash (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "deepseek-ai/deepseek-v4-pro", Name: "DeepSeek V4 Pro (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "deepseek-ai/deepseek-r1", Name: "DeepSeek R1 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192, CanReason: true},
				{ID: "google/gemma-2-9b-it", Name: "Gemma 2 9B IT (NIM)", ContextWindow: 8192, DefaultMaxTokens: 8192},
				{ID: "google/gemma-3-27b-it", Name: "Gemma 3 27B IT (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "google/gemma-3n-e2b-it", Name: "Gemma 3n E2B IT (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "google/gemma-3n-e4b-it", Name: "Gemma 3n E4B IT (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "ibm/granite-3_3-8b-instruct", Name: "Granite 3.3 8B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "meta/llama-3.1-8b-instruct", Name: "Llama 3.1 8B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "meta/llama-3.1-70b-instruct", Name: "Llama 3.1 70B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "meta/llama-3.1-405b-instruct", Name: "Llama 3.1 405B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "meta/llama-3.3-70b-instruct", Name: "Llama 3.3 70B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "microsoft/phi-3-mini-128k-instruct", Name: "Phi-3 Mini 128K Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "microsoft/phi-4-mini-instruct", Name: "Phi-4 Mini Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "minimaxai/minimax-m2.7", Name: "MiniMax M2.7 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "mistralai/mistral-7b-instruct", Name: "Mistral 7B Instruct (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "mistralai/mistral-7b-instruct-v0.3", Name: "Mistral 7B Instruct v0.3 (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "mistralai/mistral-large", Name: "Mistral Large (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "mistralai/mixtral-8x7b-instruct", Name: "Mixtral 8x7B Instruct (NIM)", ContextWindow: 32768, DefaultMaxTokens: 8192},
				{ID: "mistralai/mixtral-8x22b-instruct", Name: "Mixtral 8x22B Instruct (NIM)", ContextWindow: 65536, DefaultMaxTokens: 8192},
				{ID: "moonshotai/kimi-k2-instruct", Name: "Kimi K2 Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "moonshotai/kimi-k2-instruct-0905", Name: "Kimi K2 Instruct 0905 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "moonshotai/kimi-k2-thinking", Name: "Kimi K2 Thinking (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192, CanReason: true},
				{ID: "nvidia/llama3-chatqa-1.5-70b", Name: "Llama3 ChatQA 1.5 70B (NIM)", ContextWindow: 128000, DefaultMaxTokens: 4096},
				{ID: "nvidia/nemotron-4-340b-instruct", Name: "Nemotron 4 340B Instruct (NIM)", ContextWindow: 4096, DefaultMaxTokens: 4096},
				{ID: "qwen/qwen2-7b-instruct", Name: "Qwen 2 7B Instruct (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "qwen/qwen2.5-7b-instruct", Name: "Qwen 2.5 7B Instruct (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "qwen/qwen2.5-coder-7b-instruct", Name: "Qwen 2.5 Coder 7B Instruct (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "qwen/qwen3-5-122b-a10b", Name: "Qwen 3.5 122B A10B (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "qwen/qwen3-coder-480b-a35b-instruct", Name: "Qwen 3 Coder 480B A35B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "qwen/qwen3-next-80b-a3b-instruct", Name: "Qwen 3 Next 80B A3B Instruct (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192},
				{ID: "qwen/qwen3-next-80b-a3b-thinking", Name: "Qwen 3 Next 80B A3B Thinking (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192, CanReason: true},
				{ID: "qwen/qwq-32b", Name: "QwQ 32B (NIM)", ContextWindow: 131072, DefaultMaxTokens: 8192, CanReason: true},
				{ID: "seallms/seallm-7b-v2.5", Name: "SeaLLM 7B v2.5 (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "upstage/solar-10.7b-instruct", Name: "Solar 10.7B Instruct (NIM)", ContextWindow: 32768, DefaultMaxTokens: 4096},
				{ID: "z-ai/glm4.7", Name: "GLM 4.7 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192, CanReason: true},
				{ID: "z-ai/glm5", Name: "GLM 5 (NIM)", ContextWindow: 128000, DefaultMaxTokens: 8192, CanReason: true},
			},
		},
	}
}
