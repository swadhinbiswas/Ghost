// ollama_free.go: Ollama local provider (truly free, runs on user's GPU).
//
// Base URL: http://localhost:11434/v1
// Auth: none
//
// User runs `ollama serve` then `ollama pull <model>` once.
package providers

import "charm.land/catwalk/pkg/catwalk"

var OllamaLocalProvider = catwalk.Provider{
	ID:                  "ollama-local",
	Name:                "Ollama (local, free forever)",
	APIEndpoint:         "http://localhost:11434/v1",
	APIKey:              "ollama", // required by some clients but ignored by Ollama
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "qwen2.5-coder:32b",
	DefaultSmallModelID: "qwen2.5-coder:7b",
	Models: []catwalk.Model{
		{ID: "qwen2.5-coder:32b", Name: "Qwen 2.5 Coder 32B (Ollama, local)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "qwen2.5-coder:14b", Name: "Qwen 2.5 Coder 14B (Ollama, local)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "qwen2.5-coder:7b", Name: "Qwen 2.5 Coder 7B (Ollama, local)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "qwen3-coder:30b", Name: "Qwen 3 Coder 30B (Ollama, local)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "deepseek-coder-v2:16b", Name: "DeepSeek Coder V2 16B (Ollama)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "codellama:34b", Name: "CodeLlama 34B (Ollama)", ContextWindow: 100_000, DefaultMaxTokens: 8_192},
		{ID: "codellama:13b", Name: "CodeLlama 13B (Ollama)", ContextWindow: 16_384, DefaultMaxTokens: 8_192},
		{ID: "codellama:7b", Name: "CodeLlama 7B (Ollama)", ContextWindow: 16_384, DefaultMaxTokens: 8_192},
		{ID: "llama3.3:70b", Name: "Llama 3.3 70B (Ollama)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "llama3.2:latest", Name: "Llama 3.2 Latest (Ollama)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "mistral:latest", Name: "Mistral Latest (Ollama)", ContextWindow: 32_768, DefaultMaxTokens: 8_192},
		{ID: "mixtral:8x22b", Name: "Mixtral 8x22B (Ollama)", ContextWindow: 64_000, DefaultMaxTokens: 8_192},
		{ID: "gemma2:27b", Name: "Gemma 2 27B (Ollama)", ContextWindow: 8_192, DefaultMaxTokens: 8_192},
		{ID: "phi3:medium", Name: "Phi-3 Medium (Ollama)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "command-r", Name: "Command R (Ollama)", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "starcoder2:15b", Name: "StarCoder2 15B (Ollama)", ContextWindow: 16_384, DefaultMaxTokens: 8_192},
	},
}
