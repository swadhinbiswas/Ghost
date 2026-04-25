# Ghost CLI - Comprehensive Architecture Analysis & Enhancement Plan

## Executive Summary

Ghost is a sophisticated Go-based CLI for AI-assisted coding, built on top of Charm's BubbleTea v2 framework. The current architecture is functionally solid but faces critical design challenges in UI/UX, code organization, and scalability. This document provides a complete analysis of the existing system and a phased roadmap for transforming Ghost into a best-in-class terminal AI assistant.

---

## Part 1: Current Architecture Analysis

### 1.1 Core System Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         Ghost CLI Application                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   TUI Layer  │    │ Agent Layer  │    │ Data Layer   │      │
│  │  (BubbleTea) │    │ (LLM Agent)  │    │  (SQLite)    │      │
│  └───────┬──────┘    └──────┬───────┘    └──────┬───────┘      │
│          │                  │                   │               │
│  ┌───────▼──────────────────▼───────────────────▼────────────┐  │
│  │            Service Integration Layer                      │  │
│  │  ┌─────────┬─────────┬──────────┬──────────┬──────────┐   │  │
│  │  │Session  │Message  │History   │Permission│FileTrack │   │  │
│  │  │Service  │Service  │Service   │Service   │Service   │   │  │
│  │  └─────────┴─────────┴──────────┴──────────┴──────────┘   │  │
│  └──────────────────────────────────────────────────────────┘  │
│          │                                                      │
│  ┌───────▼──────────────────────────────────────────────────┐  │
│  │  Tool Execution & Provider Integration                  │  │
│  │  ┌──────────┬──────────┬─────────┬──────────────┐       │  │
│  │  │Bash Tool │Edit Tool │Web Fetch│Memory Tool  │ ...   │  │
│  │  └──────────┴──────────┴─────────┴──────────────┘       │  │
│  │  ┌──────────┬──────────┬─────────┬──────────────┐       │  │
│  │  │ OpenAI  │ Anthropic │ Google  │ Other LLMs  │       │  │
│  │  └──────────┴──────────┴─────────┴──────────────┘       │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Directory Structure Breakdown

```
internal/
├── agent/              # Core LLM orchestration & execution
│   ├── agent.go        # Main agent coordination logic
│   ├── prompt/         # System prompt templating
│   ├── templates/      # Prompt templates for title, summary, coder
│   ├── tools/          # 25+ executable tools (bash, edit, read, etc.)
│   ├── loop.go        # Agentic loop detection & management
│   └── notify/         # Notification broker for tool execution
├── app/               # Application lifecycle & service wiring
│   └── app.go         # Main App struct coordinating all services
├── config/            # Configuration management
│   └── config.go      # Multi-provider LLM config, permissions
├── db/                # Database & ORM
│   ├── migrations/     # SQL migration files
│   └── queries/        # SQLC generated helpers
├── session/           # Session management
├── message/           # Message storage & retrieval
├── history/           # File history tracking
├── ui/                # Terminal UI components
│   ├── model/         # BubbleTea main model (2200+ lines)
│   ├── chat/          # Message rendering (user, assistant, tools)
│   ├── styles/        # Themed styles & colors
│   ├── dialog/        # Dialog overlays
│   └── anim/          # Animations & spinners
├── llm/               # LLM provider abstraction (Catwalk)
├── tools/             # External tool support (bash execution, etc.)
├── permission/        # Permission system for tool execution
└── ... (other services)
```

### 1.3 Data Flow Architecture

```
User Input (Textarea/Command)
    ↓
[UI Model] - Receive KeyMsg / TextMsg
    ↓
Validate & Format Input
    ↓
[Session Service] - Store user message as Message
    ↓
[Agent Coordinator] - Queue prompt for LLM
    ↓
[LLM Provider] (OpenAI/Anthropic/etc)
    ↓
Receive Response + Tool Calls
    ↓
[Tool Executor] - Execute bash, file edits, searches
    ↓
Collect Tool Results
    ↓
[Agent Loop] - Embed tool results in next prompt
    ↓
Final Response Ready
    ↓
[UI Chat Renderer] - Display message stream + spinners
    ↓
Store in [Message Service]
    ↓
Update UI State → Ready for next input
```

### 1.4 Key Technologies & Dependencies

| Layer         | Technology                                    | Purpose                              |
|---------------|-----------------------------------------------|--------------------------------------|
| TUI Framework | BubbleTea v2 + Lipgloss v2                   | Terminal UI rendering               |
| Current Issue | **Ultraviolet** (fullscreen canvas rendering) | **ARCHITECTURAL BLOCKER**            |
| Styling       | Lipgloss + Charm color profiles              | Terminal styling & themes           |
| LLM Abstraction | Catwalk (multi-provider)                    | Unified LLM interface               |
| Database      | SQLite3 (ncruces/go-sqlite3)                 | Session/message persistence         |
| Shell Execution | Shell (bash/sh)                             | Tool execution & scripting          |
| Markdown      | Glamour v2 + Chroma v2                       | Code highlighting & markdown render |
| MCP Support   | MCP Go SDK                                    | Model Context Protocol integration  |
| Git Support   | go-git/v5                                    | Git operations & diffs              |

---

## Part 2: Current Architecture Strengths & Weaknesses

### 2.1 Strengths ✅

1. **Solid Service Layer:** Well-organized service interfaces (Session, Message, History, Permission, FileTracker)
2. **Multi-Provider LLM Support:** Catwalk abstraction allows seamless provider switching
3. **Comprehensive Tooling:** 25+ built-in tools covering bash, file ops, search, web fetch, MCP
4. **Memory System:** Project-level persistent memory (`.ghost/memory.json`) for context preservation
5. **Agentic Loops:** Sophisticated loop detection and management
6. **Configuration:** Flexible config system supporting custom providers and model overrides
7. **Database Design:** Clean schema with migrations and SQLC type-safe generation

### 2.2 Critical Weaknesses ❌

#### 2.2.1 UI/UX Issues

| Issue                     | Current State                                     | Impact                                  |
|---------------------------|---------------------------------------------------|-----------------------------------------|
| **Fullscreen Layout**     | Uses `ultraviolet` for absolute positioned canvas | Blocks linear/scrolling terminal output |
| **Tool Visualization**    | Large, blocky card-based UI with animations      | Cluttered, hard to scan quickly         |
| **No Output Streaming**   | Chat history confined to windowed viewport       | Can't scroll naturally in terminal      |
| **Complex State**         | 2200+ line UI model with multiple layout systems | Hard to maintain, fix, debug            |
| **Keyboard Navigation**   | Focus state management is convoluted             | Mouse dependency for complex UIs        |

#### 2.2.2 Code Organization Issues

| Issue                     | Current State                                   | Impact                                      |
|---------------------------|--------------------------------------------------|---------------------------------------------|
| **Monolithic UI Model**   | Single `ui.go` file with 2200+ lines            | Single point of failure, hard to test       |
| **Chat Rendering Logic** | Scattered across multiple chat/*.go files       | No single source of truth for message format|
| **Tool Output Rendering** | Tightly coupled to UI model                      | Can't reuse tool rendering logic elsewhere  |
| **Styles Hardcoded**      | Theme colors scattered across styles package    | No real theme switching capability          |
| **Layout System**         | Custom `uiLayout` struct + ultraviolet overlay  | Not following BubbleTea patterns            |

#### 2.2.3 Scalability & Maintenance

| Issue                          | Current State                              | Impact                                    |
|--------------------------------|--------------------------------------------|--------------------------------------------|
| **No Streaming Optimization**  | All content rendered at once               | Latency issues with large conversations    |
| **Limited Extensibility**      | Tools hardcoded in `internal/agent/tools/` | Hard to add custom user-defined tools      |
| **Incomplete Test Coverage**   | Tools lack table-driven unit tests         | Risk of silent bugs in file mutations      |
| **Missing Documentation**      | No runbook for new contributors            | Steep learning curve for onboarding        |
| **Theme System**               | No abstraction for theme switching         | Can't support light/dark mode elegantly    |

---

## Part 3: Critical Design Gaps

### 3.1 Gap 1: Linear/Scrolling UI vs. Fullscreen IDE

**Current Problem:**
- Ultraviolet forces a fullscreen, absolute-positioned layout like an IDE
- Chat history is confined to a viewport window
- Can't naturally scroll up to see old messages (cursor instead) with traditional terminal scrolling
- Doesn't follow Claude Code's elegant linear "print to stdout, then live prompt zone" pattern

**User Requirement:**
- Linear scrolling: Chat messages appear as natural terminal lines
- Minimal tools: Single-line spinners → green checkmarks
- Natural terminal experience: Works with screen multiplexers (tmux, vim)

**Why It's Important:**
The UI paradigm fundamentally shapes the user experience. A fullscreen IDE-like interface contradicts the Unix philosophy of "composable tools."

---

### 3.2 Gap 2: Tool Execution Visualization

**Current Problem:**
- Tools show as large, animated cards in the viewport
- Multiple tools create visual clutter
- No clear, scannable "what's happening" status
- Takes up significant screen real estate

**Desired Pattern:**
```
Ghost is analyzing src/handler.ts...
[✓] Analyzed handler.ts (15 lines)
[✓] Identified 3 HTTP endpoints
[✓] Generated test suite (45 lines)

Now let's write the tests:
---
```

**Impact:**
- Cleaner terminal output
- Better integration with pipes, redirects, and tmux
- Matches user expectations from other CLI tools

---

### 3.3 Gap 3: Theme & Branding System

**Current Problem:**
- Styles are defined as static `lipgloss.Style` instances
- Primary/Secondary colors are hardcoded
- No easy way to switch themes or support light/dark mode
- User branding is not possible

**Needed Infrastructure:**
- Theme abstraction layer
- Color palette definitions
- Runtime theme switching support
- User-custom theme files

---

### 3.4 Gap 4: Code Organization & Testability

**Current Problem:**
- UI model is monolithic (2200+ lines)
- Chat rendering mixed with state management
- Tool execution tightly coupled to UI
- Hard to unit test individual components

**Refactoring Needed:**
- Split UI model into smaller, composable components
- Extract rendering logic into pure functions
- Create tool rendering abstraction
- Build comprehensive test suite

---

### 3.5 Gap 5: Extensibility & Plugin Architecture

**Current Problem:**
- Tools are hardcoded in the binary
- No way for users to define custom tools without forking
- MCP support exists but isn't well-integrated into UI
- Skills system is rudimentary

**Opportunity:**
- Define a plugin interface
- Support loading external tool binaries
- Create a marketplace for community tools
- Document plugin development

---

## Part 4: Enhanced Architecture Proposal

### 4.1 Phase 1: UI/UX Modernization (High Priority)

#### Objective: Replace Ultraviolet fullscreen layout with linear scrolling BubbleTea output

**Key Changes:**

1. **Remove Ultraviolet Dependency**
   ```go
   // Current (BAD):
   canvas := uv.NewScreenBuffer(m.width, m.height)
   v.Cursor = m.Draw(canvas, canvas.Bounds())
   content := canvas.Render()
   
   // Proposed (GOOD):
   content := m.renderScrollingLayout()  // Returns plain string
   v.Content = content
   ```

2. **Implement Linear Output Renderer**
   ```go
   type ScrollingRenderer struct {
       chatHistory    []RenderedMessage
       inputPrompt    string
       toolSpinners   map[string]*spinner.Model
       maxHistorySize int
   }
   
   func (r *ScrollingRenderer) Render() string {
       var buf strings.Builder
       // Render historical messages
       for _, msg := range r.chatHistory {
           buf.WriteString(msg.String())
           buf.WriteString("\n\n")
       }
       // Render active spinners (tools running)
       for _, spinner := range r.toolSpinners {
           buf.WriteString(spinner.View())
           buf.WriteString("\n")
       }
       // Render input prompt
       buf.WriteString(r.inputPrompt)
       return buf.String()
   }
   ```

3. **Minimize Tool Visualization**
   ```
   Before:
   ╔════════════════════════════════════════╗
   ║  🔧 Edit - src/handler.ts              ║
   ║  ─────────────────────────────────────  ║
   ║  Modified 14 lines across 3 sections   ║
   ║  Status: ✓ Complete (150ms)            ║
   ╚════════════════════════════════════════╝
   
   After:
   [✓] Modified src/handler.ts (14 lines)
   ```

4. **Disable AltScreen Mode**
   ```go
   v := tea.NewView("")
   v.AltScreen = false  // Allow natural scrolling
   ```

#### Implementation Steps:
1. Create `internal/ui/renderer/` package with `ScrollingRenderer`
2. Refactor `ui.go` to use `ScrollingRenderer` instead of `ultraviolet`
3. Update tool rendering in `internal/ui/chat/` for single-line output
4. Test with various message lengths and terminal sizes
5. Remove ultraviolet dependency from `go.mod`

---

### 4.2 Phase 2: Component Decomposition & Testability

#### Objective: Break monolithic UI model into focused, composable components

**New Directory Structure:**
```
internal/ui/
├── model/
│   ├── model.go              # Core BubbleTea Model interface
│   ├── session_state.go      # Session/conversation state
│   └── input_handler.go      # Input processing logic
├── renderer/
│   ├── layout.go             # Linear layout rendering
│   ├── message.go            # Message formatting
│   └── tool_output.go        # Tool result formatting
├── components/
│   ├── textarea_input.go     # Input component
│   ├── chat_history.go       # Chat history component
│   ├── tool_status.go        # Tool status display
│   └── help_panel.go         # Help/keybindings panel
├── theme/
│   ├── theme.go              # Theme abstraction
│   ├── colors.go             # Color palette definitions
│   └── builtin.go            # Built-in themes
└── ...
```

**Component APIs:**
```go
// Message Renderer
type MessageRenderer interface {
    Render(msg *message.Message, width int) string
}

// Tool Renderer
type ToolRenderer interface {
    RenderPending(toolName string) string
    RenderSuccess(toolName string, duration time.Duration) string
    RenderError(toolName string, err error) string
}

// Theme Manager
type ThemeManager interface {
    GetTheme(name string) *Theme
    SetTheme(name string) error
    ListThemes() []string
}
```

#### Benefits:
✅ Easier unit testing
✅ Reusable rendering logic
✅ Clear separation of concerns
✅ Easier to modify behavior without side effects
✅ Better documentation

---

### 4.3 Phase 3: Theme & Branding System

#### Objective: Implement comprehensive, user-customizable theming

**Theme Definition:**
```yaml
# ~/.ghost/themes/my-theme.yaml
name: "My Custom Ghost"
description: "Professional dark theme"

colors:
  primary: "#00D9FF"      # Cyan
  secondary: "#FF006E"    # Pink
  success: "#06FFA5"      # Green
  error: "#FF006E"        # Red
  warning: "#FFB703"      # Orange
  muted: "#606060"        # Dark gray
  
  background: "#0A0E27"   # Very dark blue
  foreground: "#E0E0E0"   # Light gray

styles:
  message_user:
    bold: true
    color: "primary"
  message_assistant:
    color: "foreground"
  tool_success:
    bold: true
    color: "success"
  tool_error:
    bold: true
    color: "error"
```

**Implementation:**
```go
type Theme struct {
    Name        string
    Description string
    Colors      map[string]string
    Styles      map[string]StyleConfig
}

type StyleConfig struct {
    Bold      bool
    Italic    bool
    Underline bool
    Color     string
    BgColor   string
}

type ThemeManager struct {
    builtins    map[string]*Theme
    userThemes  map[string]*Theme
    currentName string
}
```

---

### 4.4 Phase 4: Extensibility & Plugin System

#### Objective: Allow users to define custom tools and extend Ghost

**Plugin Interface:**
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args ...string) (string, error)
    Validate(args ...string) error
}

type ToolPlugin struct {
    Binary string  // Path to executable
    Args   []string
}

func (tp *ToolPlugin) Execute(ctx context.Context, args ...string) (string, error) {
    cmd := exec.CommandContext(ctx, tp.Binary, append(tp.Args, args...)...)
    return cmd.CombinedOutput()
}
```

**Plugin Discovery:**
```
~/.ghost/plugins/
├── custom_lint/
│   ├── tool.toml
│   └── bin/custom-lint
├── jira_search/
│   ├── tool.toml
│   └── bin/jira-search
└── ...
```

**Tool Discovery:**
```go
type PluginManager struct {
    discoveredTools map[string]Tool
}

func (pm *PluginManager) DiscoverPlugins(pluginDir string) error {
    // Scan plugin directory
    // Load tool.toml from each subdirectory
    // Register tools with agent
}
```

---

### 4.5 Phase 5: Documentation & Testing

#### Objective: Create comprehensive tests and developer documentation

**Testing Strategy:**
1. **Unit Tests:** Pure functions (renderers, formatters)
2. **Integration Tests:** Tool execution with mocked LLM
3. **E2E Tests:** Full workflow with test repositories
4. **Regression Tests:** Snapshot tests for UI output

```go
func TestMessageRenderer_User(t *testing.T) {
    cases := []struct {
        name string
        msg  *message.Message
        want string
    }{
        {
            name: "simple text",
            msg: &message.Message{
                Role: message.RoleUser,
                Content: message.Content{Text: "Hello"},
            },
            want: "You: Hello\n",
        },
        // ... more cases
    }
    
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            r := NewMessageRenderer()
            got := r.Render(tc.msg, 80)
            if got != tc.want {
                t.Errorf("got %q, want %q", got, tc.want)
            }
        })
    }
}
```

**Developer Guide Topics:**
- Architecture overview
- Component model
- Adding a new tool
- Creating a custom theme
- Building a plugin
- Writing tests
- Contributing

---

## Part 5: Implementation Roadmap

### Timeline: 8-12 Weeks

```
Week 1-2:    Phase 1 Start - Remove Ultraviolet, build ScrollingRenderer
Week 2-3:    Phase 1 - Refactor tool visualization, test scrolling output
Week 4:      Phase 1 - Polish & bug fixes
Week 5-6:    Phase 2 - Component decomposition & refactoring
Week 7:      Phase 2 - Unit test suite
Week 8-9:    Phase 3 - Theme system implementation
Week 10:     Phase 4 - Plugin architecture scaffolding
Week 11:     Phase 5 - Documentation & guides
Week 12:     Testing, polish, release
```

### Priority Matrix

```
High Impact, Low Effort:
✅ Remove Ultraviolet + Linear Rendering
✅ Minimize Tool Visualization
✅ Component Decomposition

High Impact, Medium Effort:
✅ Theme System
✅ Unit Tests

Medium Impact, Medium Effort:
✅ Plugin System
✅ Developer Documentation

Lower Priority (Future):
⚪ RAG/Vector Search
⚪ Docker Sandboxing
⚪ AST-Aware Editing
```

---

## Part 6: Success Metrics

### Quantitative Metrics
- ✅ Reduce UI model from 2200 lines → <500 lines (main model)
- ✅ Increase test coverage from ~20% → >80%
- ✅ Support 3+ built-in themes
- ✅ <50ms render time for 100-message conversation
- ✅ Support custom plugins (minimum 1 example plugin)

### Qualitative Metrics
- ✅ "Ghost feels like a natural terminal tool" (user feedback)
- ✅ Code is approachable for new contributors
- ✅ Rendering logic is testable and understandable
- ✅ Easy to customize appearance and behavior

---

## Part 7: Risk Mitigation

| Risk                              | Likelihood | Severity | Mitigation                                |
|-----------------------------------|------------|----------|-------------------------------------------|
| Ultraviolet removal breaks UI     | Medium     | High     | Use feature branch, comprehensive testing |
| Component decomposition is messy  | Medium     | Medium   | Start with clear interfaces               |
| Theme system is too complex       | Low        | Medium   | Keep initial version simple, extend later |
| Plugin system has security issues | High       | High     | Sandboxing, signature verification        |
| Performance regression            | Low        | High     | Benchmark before & after                 |

---

## Part 8: Conclusion & Next Steps

Ghost has a solid technical foundation but suffers from UI/UX and code organization challenges. The proposed phased refactoring addresses these systematically while maintaining backward compatibility where possible.

### Immediate Actions (Next Sprint):

1. **Create feature branch:** `refactor/ui-modernization`
2. **Build ScrollingRenderer prototype** in Phase 1
3. **Create architecture ADR** (Architecture Decision Record)
4. **Establish component interfaces** for testing
5. **Set up performance benchmarks** to track regressions

### Success = Perfect Polish

With this roadmap executed, Ghost will evolve from a "solid clone" to a **best-in-class, terminal-native AI coding assistant** that feels natural, responds quickly, and invites contribution from the community.

---

**Document Version:** 1.0  
**Last Updated:** 2026-04-01  
**Status:** Ready for Implementation
