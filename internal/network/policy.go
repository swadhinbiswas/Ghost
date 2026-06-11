// Package network centralises Ghost's network-egress policy.
//
// Stealth mode (env GHOST_STEALTH=1) blocks every outbound call except those
// explicitly allow-listed for the user's selected model API.
//
// Default: telemetry is opt-in (GHOST_OPT_IN_METRICS=1). Stealth flips that off
// and additionally blocks update checks, catwalk fetches, MCP discovery, and
// the embedder background indexer.
package network

import (
	"os"
	"sync/atomic"
)

var (
	stealth atomic.Bool
	optedIn atomic.Bool // explicit telemetry opt-in
)

func init() {
	if os.Getenv("GHOST_STEALTH") == "1" {
		stealth.Store(true)
	}
	if os.Getenv("GHOST_OPT_IN_METRICS") == "1" {
		optedIn.Store(true)
	}
}

// Stealth returns true if the user has enabled stealth mode.
func Stealth() bool { return stealth.Load() }

// OptedIn returns true if the user has explicitly opted in to telemetry.
func OptedIn() bool { return optedIn.Load() }

// Telemetry returns true if telemetry should actually be sent (opted in AND
// not in stealth mode).
func Telemetry() bool { return OptedIn() && !Stealth() }

// SetStealth is for tests and the --stealth CLI flag.
func SetStealth(v bool) { stealth.Store(v) }

// SetOptedIn is for tests and --opt-in-metrics.
func SetOptedIn(v bool) { optedIn.Store(v) }

// allowListHosts is the set of hostnames that are *always* allowed (model APIs
// the user explicitly opted into). Compared case-insensitively against the
// request's URL.Host.
var allowListHosts = map[string]struct{}{
	// Local / on-device
	"localhost": {},
	"127.0.0.1": {},
	"::1":       {},
	// OpenAI
	"api.openai.com": {},
	// Anthropic
	"api.anthropic.com": {},
	// Google
	"generativelanguage.googleapis.com": {},
	"aiplatform.googleapis.com":         {},
	// Mistral
	"api.mistral.ai": {},
	// Groq
	"api.groq.com": {},
	// Cerebras
	"api.cerebras.ai": {},
	// Cohere
	"api.cohere.com": {},
	// xAI
	"api.x.ai": {},
	// OpenRouter
	"openrouter.ai": {},
	// 9Router meta-router (local)
	// already covered by localhost
	// Kiro
	"api.kiro.ai": {},
	// OpenCode Zen
	"opencode.ai": {},
	// NVIDIA NIM
	"integrate.api.nvidia.com": {},
	// GitHub
	"api.githubcopilot.com": {},
	"models.github.ai":      {},
	"api.github.com":        {},
	// Qwen / Alibaba
	"dashscope.aliyuncs.com": {},
	// iFlow
	"api.iflow.cn": {},
	// Cloudflare
	"api.cloudflare.com": {},
	// Hugging Face
	"api-inference.huggingface.co": {},
	// Ollama (local)
	// already covered by localhost
	// Hyperbolic, Together, DeepSeek, Fireworks, etc.
	"api.together.xyz":   {},
	"api.deepseek.com":   {},
	"api.fireworks.ai":   {},
	"api.hyperbolic.xyz": {},
}

// Allow returns true if a request to host is permitted under current policy.
func Allow(host string) bool {
	if Stealth() {
		_, ok := allowListHosts[host]
		return ok
	}
	return true
}
