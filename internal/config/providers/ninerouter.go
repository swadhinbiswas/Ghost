// ninerouter.go: 9Router meta-router provider.
//
// 9Router is a local OpenAI-compatible proxy (https://github.com/decolua/9router)
// that fans out to 60+ providers with smart 3-tier fallback (subscription ->
// cheap -> free). Once the user runs `npm install -g 9router` and starts the
// daemon, every model listed below becomes reachable at:
//
//	http://localhost:20128/v1
//
// Default API key: "sk_9router" (no real auth on localhost).
//
// Source: https://github.com/decolua/9router README
package providers

import "charm.land/catwalk/pkg/catwalk"

// NineRouterProvider exposes 9Router's curated model set.
//
// All model IDs are passed through verbatim. When 9Router is running, the
// model list is enriched dynamically from GET /v1/models.
var NineRouterProvider = catwalk.Provider{
	ID:                  "9router",
	Name:                "9Router (60+ providers, free tier)",
	APIEndpoint:         "http://localhost:20128/v1",
	APIKey:              "$NINEROUTER_API_KEY",
	Type:                catwalk.TypeOpenAICompat,
	DefaultLargeModelID: "kr/claude-sonnet-4.5",
	DefaultSmallModelID: "oc/gpt-5-nano",
	Models: []catwalk.Model{
		// Kiro AI tier (free, unlimited) - kr/ prefix
		{ID: "kr/claude-sonnet-4-5", Name: "Kiro Claude Sonnet 4.5 (free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "kr/claude-haiku-4-5", Name: "Kiro Claude Haiku 4.5 (free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192},
		{ID: "kr/glm-5", Name: "Kiro GLM-5 (free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "kr/MiniMax-M2.5", Name: "Kiro MiniMax M2.5 (free)", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "kr/qwen3-coder-next", Name: "Kiro Qwen3 Coder Next (free)", ContextWindow: 262_144, DefaultMaxTokens: 8_192},
		{ID: "kr/deepseek-3.2", Name: "Kiro DeepSeek 3.2 (free)", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CanReason: true},
		// OpenCode Zen tier (free) - oc/ prefix
		{ID: "oc/big-pickle", Name: "OpenCode Zen Big Pickle (free)", ContextWindow: 200_000, DefaultMaxTokens: 8_192},
		{ID: "oc/deepseek-v4-flash-free", Name: "OpenCode Zen DeepSeek V4 Flash Free", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "oc/mimo-v2.5-free", Name: "OpenCode Zen MiMo V2.5 Free", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "oc/qwen3.6-plus-free", Name: "OpenCode Zen Qwen 3.6 Plus Free", ContextWindow: 128_000, DefaultMaxTokens: 8_192},
		{ID: "oc/minimax-m3-free", Name: "OpenCode Zen MiniMax M3 Free", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192},
		{ID: "oc/nemotron-3-super-free", Name: "OpenCode Zen Nemotron 3 Super Free", ContextWindow: 1_000_000, DefaultMaxTokens: 8_192, CanReason: true},
		{ID: "oc/gpt-5-nano", Name: "OpenCode Zen GPT-5 Nano (cheap)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 0.05, CostPer1MOut: 0.40, CostPer1MInCached: 0.005},
		{ID: "oc/glm-5.1", Name: "OpenCode Zen GLM 5.1 (cheap)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CanReason: true, CostPer1MIn: 1.40, CostPer1MOut: 4.40, CostPer1MInCached: 0.26},
		// Claude Code subscription - cc/ prefix
		{ID: "cc/claude-opus-4-7", Name: "Claude Code Opus 4.7 (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 32_000, CanReason: true, CostPer1MIn: 5.00, CostPer1MOut: 25.00, CostPer1MInCached: 0.50, CostPer1MOutCached: 6.25},
		{ID: "cc/claude-opus-4-6", Name: "Claude Code Opus 4.6 (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 32_000, CanReason: true, CostPer1MIn: 5.00, CostPer1MOut: 25.00, CostPer1MInCached: 0.50, CostPer1MOutCached: 6.25},
		{ID: "cc/claude-sonnet-4-6", Name: "Claude Code Sonnet 4.6 (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 3.00, CostPer1MOut: 15.00, CostPer1MInCached: 0.30, CostPer1MOutCached: 3.75},
		{ID: "cc/claude-sonnet-4-5", Name: "Claude Code Sonnet 4.5 (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 3.00, CostPer1MOut: 15.00, CostPer1MInCached: 0.30, CostPer1MOutCached: 3.75},
		{ID: "cc/claude-haiku-4-5", Name: "Claude Code Haiku 4.5 (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 8_000, CostPer1MIn: 1.00, CostPer1MOut: 5.00, CostPer1MInCached: 0.10, CostPer1MOutCached: 1.25},
		// OpenAI Codex subscription - cx/ prefix
		{ID: "cx/gpt-5.5", Name: "Codex GPT 5.5 (subscription)", ContextWindow: 272_000, DefaultMaxTokens: 32_000, CostPer1MIn: 5.00, CostPer1MOut: 30.00, CostPer1MInCached: 0.50},
		{ID: "cx/gpt-5.4", Name: "Codex GPT 5.4 (subscription)", ContextWindow: 272_000, DefaultMaxTokens: 16_000, CostPer1MIn: 2.50, CostPer1MOut: 15.00, CostPer1MInCached: 0.25},
		{ID: "cx/gpt-5.3-codex", Name: "Codex GPT 5.3 Codex (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 1.75, CostPer1MOut: 14.00, CostPer1MInCached: 0.175},
		{ID: "cx/gpt-5.2-codex", Name: "Codex GPT 5.2 Codex (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 1.75, CostPer1MOut: 14.00, CostPer1MInCached: 0.175},
		{ID: "cx/gpt-5.1-codex-max", Name: "Codex GPT 5.1 Codex Max (subscription)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 1.25, CostPer1MOut: 10.00, CostPer1MInCached: 0.125},
		// GitHub Copilot - gh/ prefix
		{ID: "gh/gpt-5.4", Name: "GitHub Copilot GPT 5.4 (free quota)", ContextWindow: 272_000, DefaultMaxTokens: 16_000, CostPer1MIn: 2.50, CostPer1MOut: 15.00, CostPer1MInCached: 0.25},
		{ID: "gh/claude-opus-4.7", Name: "GitHub Copilot Claude Opus 4.7", ContextWindow: 200_000, DefaultMaxTokens: 32_000, CanReason: true, CostPer1MIn: 5.00, CostPer1MOut: 25.00, CostPer1MInCached: 0.50, CostPer1MOutCached: 6.25},
		{ID: "gh/claude-sonnet-4.6", Name: "GitHub Copilot Claude Sonnet 4.6", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 3.00, CostPer1MOut: 15.00, CostPer1MInCached: 0.30, CostPer1MOutCached: 3.75},
		{ID: "gh/gemini-3.1-pro-preview", Name: "GitHub Copilot Gemini 3.1 Pro Preview", ContextWindow: 1_000_000, DefaultMaxTokens: 16_000, CostPer1MIn: 2.00, CostPer1MOut: 12.00, CostPer1MInCached: 0.20},
		{ID: "gh/grok-code-fast-1", Name: "GitHub Copilot Grok Code Fast (free)", ContextWindow: 131_072, DefaultMaxTokens: 8_192, CostPer1MIn: 1.00, CostPer1MOut: 2.00, CostPer1MInCached: 0.20},
		// Cursor - cu/ prefix
		{ID: "cu/claude-4.6-opus-max", Name: "Cursor Claude 4.6 Opus Max", ContextWindow: 200_000, DefaultMaxTokens: 32_000, CanReason: true, CostPer1MIn: 5.00, CostPer1MOut: 25.00, CostPer1MInCached: 0.50, CostPer1MOutCached: 6.25},
		{ID: "cu/claude-4.5-sonnet-thinking", Name: "Cursor Claude 4.5 Sonnet Thinking", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CanReason: true, CostPer1MIn: 3.00, CostPer1MOut: 15.00, CostPer1MInCached: 0.30, CostPer1MOutCached: 3.75},
		{ID: "cu/gpt-5.3-codex", Name: "Cursor GPT 5.3 Codex", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CostPer1MIn: 1.75, CostPer1MOut: 14.00, CostPer1MInCached: 0.175},
		{ID: "cu/kimi-k2.5", Name: "Cursor Kimi K2.5", ContextWindow: 262_144, DefaultMaxTokens: 16_000, CostPer1MIn: 0.60, CostPer1MOut: 3.00, CostPer1MInCached: 0.10},
		// Cheap providers
		{ID: "glm/glm-5.1", Name: "GLM 5.1 (cheap)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CanReason: true, CostPer1MIn: 0.60, CostPer1MOut: 1.80, CostPer1MInCached: 0.10},
		{ID: "glm/glm-5", Name: "GLM 5 (cheap)", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CanReason: true, CostPer1MIn: 0.60, CostPer1MOut: 1.80, CostPer1MInCached: 0.10},
		{ID: "minimax/MiniMax-M2.7", Name: "MiniMax M2.7 (cheapest)", ContextWindow: 1_000_000, DefaultMaxTokens: 16_000, CostPer1MIn: 0.20, CostPer1MOut: 0.80, CostPer1MInCached: 0.04},
		{ID: "minimax/MiniMax-M2.5", Name: "MiniMax M2.5 (cheapest)", ContextWindow: 1_000_000, DefaultMaxTokens: 16_000, CostPer1MIn: 0.20, CostPer1MOut: 0.80, CostPer1MInCached: 0.04},
		{ID: "kimi/kimi-k2.5", Name: "Kimi K2.5 ($9/mo flat)", ContextWindow: 262_144, DefaultMaxTokens: 16_000, CostPer1MIn: 0.60, CostPer1MOut: 3.00, CostPer1MInCached: 0.10},
		{ID: "kimi/kimi-k2.5-thinking", Name: "Kimi K2.5 Thinking ($9/mo flat)", ContextWindow: 262_144, DefaultMaxTokens: 16_000, CanReason: true, CostPer1MIn: 0.60, CostPer1MOut: 3.00, CostPer1MInCached: 0.10},
		// Vertex AI (new GCP free credits)
		{ID: "vertex/gemini-3.1-pro-preview", Name: "Vertex Gemini 3.1 Pro Preview", ContextWindow: 1_000_000, DefaultMaxTokens: 16_000, CostPer1MIn: 2.00, CostPer1MOut: 12.00, CostPer1MInCached: 0.20},
		{ID: "vertex/gemini-3-flash-preview", Name: "Vertex Gemini 3 Flash Preview", ContextWindow: 1_000_000, DefaultMaxTokens: 16_000, CostPer1MIn: 0.50, CostPer1MOut: 3.00, CostPer1MInCached: 0.05},
		{ID: "vertex/gemini-2.5-flash", Name: "Vertex Gemini 2.5 Flash (free)", ContextWindow: 1_000_000, DefaultMaxTokens: 16_000, CostPer1MIn: 0.30, CostPer1MOut: 2.50, CostPer1MInCached: 0.03},
		{ID: "vertex-partner/glm-5-maas", Name: "Vertex Partner GLM-5 MaaS", ContextWindow: 200_000, DefaultMaxTokens: 16_000, CanReason: true, CostPer1MIn: 0.60, CostPer1MOut: 2.20, CostPer1MInCached: 0.10},
		{ID: "vertex-partner/deepseek-v3.2-maas", Name: "Vertex Partner DeepSeek V3.2 MaaS", ContextWindow: 128_000, DefaultMaxTokens: 8_192, CostPer1MIn: 0.14, CostPer1MOut: 0.28, CostPer1MInCached: 0.03},
	},
}
