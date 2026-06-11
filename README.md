<div align="center">

<img src="https://raw.githubusercontent.com/swadhinbiswas/Ghost/main/internal/cmd/stats/header.svg" width="600" alt="Ghost CLI Logo">

<p align="center">
  <br>
  <b>👻 A lightning-fast, beautiful AI assistant that lives in your terminal — powered by free models.</b>
  <br>
</p>

[![Go Release](https://img.shields.io/github/actions/workflow/status/swadhinbiswas/Ghost/release.yml?style=flat-square&logo=github&label=Release)](https://github.com/swadhinbiswas/Ghost/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-bd93f9.svg?style=flat-square&logo=opensourceinitiative)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-00ADD8?style=flat-square&logo=go)](https://golang.org/)

</div>

---

Hey there! 👋 Welcome to **Ghost** — a terminal-based AI assistant built to be fast, beautiful, and free.

Ghost ships pre-configured with generous **free** AI providers, so you can get world-class answers, code suggestions, and command help without entering a credit card. If you live in the terminal and want instant help without alt-tabbing to a browser, Ghost is for you.

> **Note on UI:** Ghost's interface is built on the foundation of [**Crush**](https://github.com/charmbracelet/crush) by [Charm](https://charm.sh/). Huge thanks to the Charm team for the gorgeous terminal toolkit.

<br>

## ✨ Why Ghost?

- **Free out of the box.** Pre-wired to free providers — no subscription, no API key required to get started.
- **Cross-platform.** A single native Go binary for Linux, macOS, and Windows (amd64 + arm64).
- **Beautiful TUI.** A responsive, themeable terminal interface (Dracula theme included).
- **Bring your own keys.** Optionally connect paid/keyed providers like Google Gemini, Groq, Cerebras, GitHub Copilot, and OpenRouter through the built-in [Catwalk](https://github.com/charmbracelet/catwalk) registry.

<br>

## 🧠 Models out of the box

These providers work with **no API key required**:

| Provider | Models | Notes |
| --- | --- | --- |
| **OpenCode Zen** | `Big Pickle`, `DeepSeek V4 Flash Free`, `Xiaomi MiMo V2.5 Free`, `Qwen 3.6 Plus Free`, `Nemotron 3 Super Free` | Free tier, no key needed |
| **OI VSCode Server** | `MiniMax M2.7 (thinking)` | Free, no key needed |

Add a key to unlock more:

| Provider | How to enable | Example models |
| --- | --- | --- |
| **Nvidia NIM** | set `NVIDIA_API_KEY` | `Nemotron 3 Ultra`, `DeepSeek V4 Pro/Flash`, `GLM 5.1`, `Kimi K2.6`, `Qwen 3.5`, `Mistral Medium 3.5`, `Step 3.7 Flash`, `MiniMax M2.7`, `DiffusionGemma` |
| **Google Gemini / Groq / Cerebras / Copilot / OpenRouter** | `ghost login` (via Catwalk registry) | provider-specific |

> Model availability is curated by each provider and may change over time.

<br>

## 🚀 Getting Started

### The fast way (macOS & Linux)

Detects your OS/architecture and installs the latest release to `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/swadhinbiswas/Ghost/main/install.sh | bash
```

### Manual download (Windows, macOS, Linux)

1. Go to the [Releases page](https://github.com/swadhinbiswas/Ghost/releases).
2. Download the archive for your system (`.tar.gz` for macOS/Linux, `.zip` for Windows).
3. Extract it and move the `ghost` binary somewhere on your `PATH`.

### Build from source

Requires **Go 1.26 or newer**:

```bash
git clone https://github.com/swadhinbiswas/Ghost.git
cd Ghost
go build -o ghost ./main.go
sudo mv ghost /usr/local/bin/
```

<br>

## 💻 Usage

Start an interactive chat session:

```bash
ghost
```

Run a one-off prompt without entering the full UI:

```bash
ghost run "Write a python script to parse a CSV file"
```

Other handy commands:

```bash
ghost free                 # pick a free model to chat with
ghost models               # list available models
ghost login                # add credentials for keyed providers
ghost theme                # switch the TUI theme
ghost update-providers     # refresh the provider registry
ghost session              # manage saved sessions
ghost --help               # see everything Ghost can do
```

<br>

## 📚 Documentation & Website

The marketing site and full documentation live in [`site/`](site/) and are built with [Astro](https://astro.build/) (using [Bun](https://bun.sh/)).

```bash
cd site
bun install
bun run dev      # local preview at http://localhost:4321
bun run build    # static build into site/dist
```

<br>

## 🛠️ Configuration

Ghost reads JSON config from standard locations (`$XDG_CONFIG_HOME/ghost/ghost.json` or `~/.config/ghost/ghost.json`). A few useful environment variables:

| Variable | Purpose |
| --- | --- |
| `GHOST_DISABLE_PROVIDER_AUTO_UPDATE` | Disable automatic provider/model refresh |
| `GHOST_DISABLE_DEFAULT_PROVIDERS` | Skip all built-in providers (BYO only) |
| `NVIDIA_API_KEY` | Enable the Nvidia NIM provider |
| `GHOST_NVIDIA_NIM_MODELS_URL` | Override the remote Nvidia NIM model catalog URL |
| `GHOST_OPENCODE_MODELS_URL` | Override the OpenCode Zen models endpoint |

On startup, Ghost refreshes the OpenCode Zen and Nvidia NIM model lists in the background (non-blocking) and caches them for the next session. This honors `GHOST_DISABLE_PROVIDER_AUTO_UPDATE`. The Nvidia NIM catalog is sourced from [`nvidia_nim_models.json`](nvidia_nim_models.json) at the repo root and fetched from its GitHub raw URL, so the model list can be updated without a new release. Run `ghost update-providers` to refresh on demand.

<br>

## 🙏 Acknowledgments

- **UI foundation:** [Crush](https://github.com/charmbracelet/crush) and the [Charm](https://charm.sh/) ecosystem (`lipgloss`, `bubbletea`, `catwalk`).
- **Free providers:** OpenCode Zen, the OI VSCode community, Nvidia, and everyone democratizing access to great AI models.

<br>

## 📄 License

Ghost is open-source software licensed under the [MIT License](LICENSE). Fork it, mod it, make it your own.
