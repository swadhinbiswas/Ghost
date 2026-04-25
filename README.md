<div align="center">

<img src="https://raw.githubusercontent.com/swadhinbiswas/ghost/main/internal/cmd/stats/header.svg" width="600" alt="Ghost CLI Logo">

<p align="center">
  <br>
  <b>A lightning-fast, highly aesthetic AI assistant living right in your terminal.</b>
  <br>
</p>

[![Go Release](https://img.shields.io/github/actions/workflow/status/swadhinbiswas/ghost/release.yml?style=flat-square&logo=github&label=Release)](https://github.com/swadhinbiswas/ghost/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square&logo=opensourceinitiative)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-00ADD8?style=flat-square&logo=go)](https://golang.org/)

</div>

---

<div align="center">
  <img src="demo.gif" alt="Ghost CLI Terminal Recording Demo" width="800">
  <br>
  <em>(See Ghost in action above!)</em>
</div>

---

Hey there! 👋 Welcome to Ghost. 

I wanted a terminal-based AI assistant that was not only incredibly fast and useful but also beautiful to look at. More importantly, I wanted to build something that runs entirely on the best **free** AI models out there. No subscriptions, no hidden API costs, just pure productivity right from your command line.

If you spend a lot of your day in the terminal and want instant answers, code suggestions, or help running commands without constantly alt-tabbing to a browser, Ghost is for you.

<br>

## <img src="https://raw.githubusercontent.com/FortAwesome/Heroicons/master/optimized/24/outline/sparkles.svg" width="24" height="24" align="bottom"> Why use Ghost?

- **It's 100% Free:** I built this to exclusively use top-tier free providers. You get access to world-class models without ever entering a credit card.
- **Works Everywhere:** It's a native Go application. Whether you're on Linux, macOS, or Windows, Ghost runs perfectly.
- **Beautiful Interface:** The terminal doesn't have to be boring. Ghost features a stunning, highly responsive UI. Huge shoutout to **Crush from Charm** for the foundational UI template that makes Ghost look so good!
- **Drop-in Proxies:** It natively connects to OpenCode Zen, GitHub Copilot, Groq, Nvidia NIM, Cerebras, and the free-tier on OpenRouter.

<br>

## <img src="https://raw.githubusercontent.com/FortAwesome/Heroicons/master/optimized/24/outline/cpu-chip.svg" width="24" height="24" align="bottom"> Models out of the box

Ghost comes pre-configured with endpoints that give you generous free tiers:

* **Google Gemini:** `Gemini 1.5 Pro` and `Gemini 1.5 Flash`
* **Nvidia NIM:** `DeepSeek R1`, `Llama 3.3 70B`, `Gemma 3`, `QwQ 32B`, and tons more!
* **Cerebras:** Instantaneous `Llama 3.1 8B` and `70B` inference.
* **GitHub Copilot:** Got a Copilot subscription? Ghost uses it as a proxy for `GPT-4o` and `Claude 3.5 Sonnet`.
* **Groq:** LPU-accelerated `Llama 3.1` and `Mixtral`.
* **OpenCode Zen:** Amazing free coding models like `Big Pickle`, `Qwen 3.6 Plus Free`, and `MiMo V2`.
* **OpenRouter:** Taps into community-provided free models like `Llama 3.1 8B Free` and `Gemma 2 9B Free`.

<br>

## <img src="https://raw.githubusercontent.com/FortAwesome/Heroicons/master/optimized/24/outline/rocket-launch.svg" width="24" height="24" align="bottom"> Getting Started

### The fast way (Mac & Linux)

Just run this one-liner in your terminal. It detects your OS and architecture and drops the latest release into `/usr/local/bin`.

```bash
curl -fsSL https://raw.githubusercontent.com/swadhinbiswas/ghost/main/install.sh | bash
```

### Manual Download (Windows, Mac, Linux)

If you prefer doing things yourself or you're on Windows:
1. Head over to the [Releases Page](https://github.com/swadhinbiswas/ghost/releases).
2. Grab the archive (`.tar.gz` for Mac/Linux, `.zip` for Windows) that matches your system.
3. Extract it and drop the `ghost` executable somewhere in your `PATH`.

### Build from source

Got Go 1.21 or higher installed? You can build it directly:

```bash
git clone https://github.com/swadhinbiswas/ghost.git
cd ghost
go build -o ghost ./main.go
sudo mv ghost /usr/local/bin/
```

<br>

## <img src="https://raw.githubusercontent.com/FortAwesome/Heroicons/master/optimized/24/outline/command-line.svg" width="24" height="24" align="bottom"> How to use it

To start an interactive chat session, just type:

```bash
ghost
```

Need to run a quick one-off command without entering the full UI? Use the `run` command:

```bash
ghost run "Write a quick python script to parse a CSV file"
```

To see everything Ghost can do:

```bash
ghost --help
```

<br>

## <img src="https://raw.githubusercontent.com/FortAwesome/Heroicons/master/optimized/24/outline/heart.svg" width="24" height="24" align="bottom"> Acknowledgments & Credits

Ghost wouldn't be possible without these amazing projects and people:

- **UI & Aesthetics:** A massive thank you to **Crush** from the [Charm](https://charm.sh/) team. Their incredible UI template gave Ghost its gorgeous look and feel. 
- **The Charm Ecosystem:** Built heavily on `lipgloss`, `bubbletea`, and `catwalk` for robust terminal rendering.
- **Free Providers:** Thank you to Google, Nvidia, Groq, Cerebras, and the OpenCode community for democratizing access to top-tier AI models.

<br>

## <img src="https://raw.githubusercontent.com/FortAwesome/Heroicons/master/optimized/24/outline/document-text.svg" width="24" height="24" align="bottom"> License

Ghost is open-source software licensed under the [MIT License](LICENSE). Feel free to fork it, mod it, and make it your own!
