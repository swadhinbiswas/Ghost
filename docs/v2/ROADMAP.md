# Ghost v2.0 — Production-Grade Stealth-Mode Roadmap

> Companion to `SKILL.md` and `AGENT.md`. This file is the **WHAT**; those two are the **HOW** for AI agents.

## Mission

Transform `swadhinbiswas/Ghost` from a Charm-Crush rebrand into a best-in-class, production-grade, terminal-first AI coding agent that:

- Ships with **every free model** anyone can find on the internet wired in by default (NVIDIA NIM, OpenCode Zen, 9Router, Kiro, OpenCode-Free, Qwen, iFlow, Groq, Cerebras, Cloudflare Workers AI, Cohere trial, Google AI Studio, GitHub Copilot, OpenRouter free).
- Runs in **stealth mode** by default (zero telemetry, zero phone-home, opt-in only).
- Matches OpenClaue's feature surface (80+ slash commands, hooks, gRPC server, bridge, per-agent routing, provider profiles, LSP, IDE integration).
- Is a single Go binary, builds on Linux/macOS/Windows, ships via Homebrew/AUR/Scoop/winget/Docker.

## Work-stream Map

| WS | Name | Critical? | Est. weeks |
|---|---|---|---|
| WS-1 | Free-model provider pack | YES | 2 |
| WS-2 | Stealth mode (zero telemetry) | YES | 1 |
| WS-3 | Slash-command expansion (10 → 70) | YES | 3 |
| WS-4 | Hook system (Pre/Post/Stop/Compact/SessionStart) | YES | 2 |
| WS-5 | gRPC server (headless mode) | YES | 2 |
| WS-6 | Bridge (remote control over WebSocket) | YES | 3 |
| WS-7 | Per-agent routing + provider profiles | YES | 2 |
| WS-8 | CI/release infra + VS Code extension | YES | 2 |
| -- | **TOTAL** | | **~17 engineer-weeks (1 person)** or **~6 weeks (3 people in parallel)** |

