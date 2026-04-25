# Ghost Agent Development Plan

To evolve `ghost` from a robust clone into a perfect, best-in-class, autonomous terminal AI assistant, we will execute the following phased development plan. This plan focuses on augmenting context visibility, ensuring safe execution, maximizing tool interoperability, and refining the developer experience.

## Phase 1: Advanced Context & Memory (Make the AI Smarter)

Currently, the agent relies on raw string search (`grep`, `ls`) and an SQLite database for chat sessions. To become a world-class assistant, it must truly "understand" the codebase.

- [ ] **Semantic Code Search (RAG):**
  - Integrate a vector database engine (e.g., `sqlite-vss`, Chroma, or a pure Go implementation).
  - Automatically chunk and embed the user's workspace in the background.
  - Implement a `semantic_search` tool so the agent can find conceptual answers (e.g., "Where is the authentication flow?").
- [ ] **AST-Aware Editing:**
  - Integrate `tree-sitter` for Go, Python, JS/TS, etc.
  - Upgrade the `edit` and `read` tools so the agent can target entire functions, classes, or interfaces by name rather than relying on brittle, exact line-number math.
- [ ] **Context Compression:**
  - Implement automatic sliding-window summarization for long chat sessions to prevent dropping valuable context when nearing API token limits.

## Phase 2: Safe & Verifiable Execution Loops

The agent has a `bash` tool. To make it a true autonomous developer, we need to let it experiment without breaking the host machine.

- [ ] **Docker Sandboxing:**
  - Introduce an interactive execution mode where shell commands and scripts are routed to a temporary Docker container rather than the host OS.
  - Mount workspace volumes safely.
- [ ] **Self-Correction Verification:**
  - When the user asks the agent to write a script or build a binary, hook a background "verify" watcher.
  - Run the resulting code silently (e.g., `go build` or `npm run build`), capture the compilation errors, feed them back into the LLM's thought loop, and return the output to the user *only* after the agent fixes its own typos.

## Phase 3: Extending the Model Context Protocol (MCP)

`ghost` already has foundational MCP support. We will expand this to connect the agent to external enterprise services.

- [ ] **Dynamic Plugin Registry:**
  - Allow users to register standard open-source MCP servers via a simple config (`~/.ghost/mcp.yaml`).
  - E.g., connect GitHub MCP (to read/write PRs from the terminal), Postgres MCP (to execute queries), or Slack MCP.
- [ ] **TUI Server Manager Panel:**
  - Add a dedicated graphical tab inside the `bubbletea` interface to monitor MCP server statuses, toggle them on/off, and read real-time connection logs.

## Phase 4: UI/UX & Reliability (The "Perfect" Polish)

- [ ] **Comprehensive Unit Testing:**
  - Most internal modules (especially in `internal/agent/tools/*`) lack test coverage.
  - Write table-driven unit tests for critical tooling (multiedits, file pathing, diff generation) to guarantee the agent modifies source code predictably.
- [ ] **Streaming TUI Optimizations:**
  - Refine the async rendering of markdown chunks over `lipgloss` and `glamour`.
  - Prevent flickering interface redraws during rapid, deeply-nested LLM thought generation.
- [ ] **Build & Release Pipelines:**
  - Set up `.github/workflows/release.yml` using `goreleaser`.
  - Automatically compile, sign, and publish `ghost` binaries across macOS, Linux, and Windows for effortless distribution via Homebrew or apt.

---

### Executing the Plan

We will tackle this incrementally. We suggest starting with **Phase 1 (AST-Aware Editing & Context Compression)**, as improving the AI's understanding immediately improves the quality of code it writes for you.
