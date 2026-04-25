# Ghost CLI - Final Project Structure (Post-Implementation)

This document shows the target directory structure after all 5 phases are complete.

## Full Directory Tree

```
Ghost/
├── README.md                          # Updated with new architecture info
├── CHANGELOG.md                       # Complete changelog
├── IMPLEMENTATION_PLAN.md             # (This implementation plan)
├── QUICK_CHECKLIST.md                 # Quick reference
├── ARCHITECTURE_ANALYSIS.md           # Architecture deep-dive
│
├── go.mod                             # Updated (no ultraviolet)
├── go.sum
├── main.go
│
├── internal/
│   ├── agent/                         # AI agent logic (unchanged)
│   │   ├── agent.go
│   │   ├── prompt/
│   │   ├── tools/
│   │   └── ...
│   │
│   ├── ui/                            # ✨ MAJOR REFACTOR
│   │   ├── model/
│   │   │   ├── model.go                   # Refactored main model (<500 lines)
│   │   │   ├── session_state.go           # Session state management
│   │   │   ├── input_handler.go           # Input processing
│   │   │   ├── ui_test.go                 # UI model tests
│   │   │   └── ui_old_backup.go           # Original backup (for reference)
│   │   │
│   │   ├── renderer/                  # ✨ NEW: Rendering abstraction
│   │   │   ├── scrolling.go                # Linear scrolling renderer
│   │   │   ├── scrolling_test.go
│   │   │   ├── message_formatter.go        # Message formatting
│   │   │   ├── message_formatter_test.go
│   │   │   ├── tool_formatter.go           # Tool output formatting
│   │   │   ├── tool_formatter_test.go
│   │   │   ├── layout.go                   # Layout composition
│   │   │   └── layout_test.go
│   │   │
│   │   ├── components/                # ✨ NEW: Composable components
│   │   │   ├── interfaces.go               # Component contracts
│   │   │   ├── input.go                    # Input component
│   │   │   ├── input_test.go
│   │   │   ├── history.go                  # Chat history component
│   │   │   ├── history_test.go
│   │   │   ├── tool_status.go              # Tool status component
│   │   │   ├── tool_status_test.go
│   │   │   ├── help.go                     # Help panel component
│   │   │   ├── help_test.go
│   │   │   └── README.md                   # Component development guide
│   │   │
│   │   ├── theme/                     # ✨ NEW: Theme system
│   │   │   ├── types.go                    # Theme type definitions
│   │   │   ├── manager.go                  # Theme manager
│   │   │   ├── manager_test.go
│   │   │   ├── builtins.go                 # Built-in themes
│   │   │   ├── builtins_test.go
│   │   │   ├── loader.go                   # YAML theme loader
│   │   │   ├── loader_test.go
│   │   │   └── README.md                   # Theme development guide
│   │   │
│   │   ├── chat/                      # Message rendering (refactored)
│   │   │   ├── messages.go                 # Updated for linear rendering
│   │   │   ├── assistant.go                # Minimal tool visualization
│   │   │   ├── tools.go                    # Tool formatters
│   │   │   └── ...
│   │   │
│   │   ├── styles/                    # Theme-aware styles
│   │   │   └── ...
│   │   │
│   │   ├── common/
│   │   ├── dialog/
│   │   ├── animations/
│   │   └── README.md                   # UI architecture guide
│   │
│   ├── plugins/                       # ✨ NEW: Plugin system
│   │   ├── interface.go                    # Plugin interface
│   │   ├── manager.go                      # Plugin manager
│   │   ├── manager_test.go
│   │   ├── loader.go                       # Plugin loader
│   │   ├── loader_test.go
│   │   ├── registry.go                     # Plugin registry
│   │   ├── sandbox.go                      # Plugin sandboxing
│   │   └── README.md                       # Plugin development guide
│   │
│   ├── config/                        # Config system (updated for themes/plugins)
│   │   └── ...
│   │
│   ├── app/
│   ├── cmd/
│   ├── db/
│   ├── session/
│   ├── message/
│   ├── llm/
│   ├── tools/
│   └── ...                             # Other modules (mostly unchanged)
│
├── cmd/
│   ├── root.go                         # Updated CLI
│   └── config.go                        # Theme/plugin config commands
│
├── examples/
│   ├── themes/                         # ✨ NEW: Theme examples
│   │   ├── dark.yaml
│   │   ├── light.yaml
│   │   ├── branded.yaml
│   │   ├── minimal.yaml
│   │   └── high-contrast.yaml
│   │
│   └── plugins/                        # ✨ NEW: Plugin examples
│       ├── hello-world/
│       │   ├── tool.toml
│       │   ├── main.go
│       │   └── README.md
│       ├── github-search/
│       │   ├── tool.toml
│       │   ├── main.go
│       │   └── README.md
│       └── jira-integration/
│           ├── tool.toml
│           ├── main.go
│           └── README.md
│
├── docs/
│   ├── README.md                       # Documentation index
│   ├── getting-started.md              # (existing)
│   │
│   ├── architecture/                   # ✨ NEW: Architecture docs
│   │   ├── overview.md                     # System architecture
│   │   ├── component-model.md              # Component architecture
│   │   ├── data-flow.md                    # Data flow diagrams
│   │   └── decision-log.md                 # Architecture decisions
│   │
│   ├── development/                    # ✨ NEW: Developer guides
│   │   ├── contributing.md                 # Contribution guidelines
│   │   ├── setup.md                        # Development setup
│   │   ├── testing.md                      # Testing guide
│   │   ├── coding-standards.md             # Code style guide
│   │   └── debugging.md                    # Debugging guide
│   │
│   ├── features/                       # ✨ NEW: Feature guides
│   │   ├── theming.md                      # Theming user guide
│   │   ├── creating-themes.md              # Theme development
│   │   ├── plugins.md                      # Plugin guide
│   │   ├── creating-plugins.md             # Plugin development
│   │   └── cli-commands.md                 # CLI reference
│   │
│   ├── api/                            # ✨ NEW: API documentation
│   │   ├── component-api.md                # Component API reference
│   │   ├── theme-api.md                    # Theme API reference
│   │   ├── plugin-api.md                   # Plugin API reference
│   │   └── renderer-api.md                 # Renderer API reference
│   │
│   ├── tutorials/                      # ✨ NEW: Tutorials
│   │   ├── first-theme.md
│   │   ├── first-plugin.md
│   │   ├── contribute-first-pr.md
│   │   └── ...
│   │
│   ├── phase1/                         # Phase 1 docs
│   │   ├── ultraviolet-removal.md
│   │   ├── scrolling-renderer.md
│   │   └── performance-benchmarks.md
│   │
│   ├── phase2/                         # Phase 2 docs
│   │   ├── component-decomposition.md
│   │   └── testability-improvements.md
│   │
│   ├── phase3/                         # Phase 3 docs
│   │   └── theme-system.md
│   │
│   ├── phase4/                         # Phase 4 docs
│   │   └── plugin-system.md
│   │
│   └── faq.md                          # Frequently asked questions
│
├── .github/
│   ├── workflows/
│   │   ├── test.yml                    # Unit & integration tests
│   │   ├── lint.yml                    # Linting & formatting
│   │   ├── benchmark.yml               # Performance benchmarks
│   │   ├── release.yml                 # Release pipeline (goreleaser)
│   │   └── docs.yml                    # Documentation build
│   │
│   └── ISSUE_TEMPLATE/
│       ├── bug.yml
│       ├── feature.yml
│       └── documentation.yml
│
├── tests/                              # Integration & E2E tests
│   ├── integration/
│   │   ├── rendering_test.go           # Rendering integration tests
│   │   ├── theme_switching_test.go     # Theme switching tests
│   │   └── component_interaction_test.go
│   │
│   ├── e2e/
│   │   ├── full_workflow_test.go       # Full user workflow
│   │   ├── plugin_loading_test.go      # Plugin system E2E
│   │   └── stability_test.go           # Long-running stability
│   │
│   └── fixtures/
│       ├── test-theme.yaml
│       ├── test-plugin/
│       └── test-sessions/
│
├── scripts/                            # Utility scripts
│   ├── benchmark.sh                    # Run benchmarks
│   ├── profile.sh                      # Profiling script
│   ├── lint.sh                         # Run linters
│   ├── format.sh                       # Auto-format code
│   ├── test.sh                         # Run test suite
│   ├── build.sh                        # Build binaries
│   ├── release.sh                      # Release script
│   └── docs-build.sh                   # Build documentation
│
└── .ghost/                             # Default theme/plugin location
    ├── themes/
    │   └── custom-theme.yaml           # User custom theme
    └── plugins/
        └── custom-plugin/              # User custom plugin
            ├── tool.toml
            └── bin/custom-plugin
```

---

## File Count Summary

### Before Implementation
```
Go Source Files:        ~150
Test Files:              ~30
Lines of Code:       150,000
Monolithic UI Code:    2,200
Test Coverage:          ~40%
```

### After Implementation
```
Go Source Files:        ~200 (+50 new files)
Test Files:              ~80 (+50 new tests)
Lines of Code:       160,000 (+10,000, but better organized)
Monolithic UI Code:      500 (-1,700 lines)
Test Coverage:           >90%
New Modules:             3 (renderer, components, theme, plugins)
```

---

## Key New Directories & Their Purpose

| Directory | Purpose | Files | Status |
|-----------|---------|-------|--------|
| `internal/ui/renderer/` | Rendering abstraction & logic | 6 | ✨ NEW |
| `internal/ui/components/` | Composable UI components | 8 | ✨ NEW |
| `internal/ui/theme/` | Theme system & management | 6 | ✨ NEW |
| `internal/plugins/` | Plugin framework | 5 | ✨ NEW |
| `docs/architecture/` | Architecture documentation | 4 | ✨ NEW |
| `docs/development/` | Developer guides | 5 | ✨ NEW |
| `docs/features/` | Feature documentation | 6 | ✨ NEW |
| `docs/api/` | API reference | 4 | ✨ NEW |
| `docs/tutorials/` | Tutorial guides | 4+ | ✨ NEW |
| `examples/themes/` | Theme examples | 5 | ✨ NEW |
| `examples/plugins/` | Plugin examples | 3+ | ✨ NEW |
| `tests/integration/` | Integration tests | 3+ | ✨ NEW |
| `tests/e2e/` | E2E tests | 3+ | ✨ NEW |

---

## Modified Directories

| Directory | Changes | Reason |
|-----------|---------|--------|
| `internal/ui/model/` | Major refactor | Component decomposition |
| `internal/ui/chat/` | Tool visualization updated | Minimal rendering |
| `internal/ui/styles/` | Theme-aware | New theme system |
| `cmd/` | New theme commands | Phase 3 implementation |
| `docs/` | Major expansion | Documentation phase |
| `.github/workflows/` | New CI/CD | Release automation |

---

## Package Dependencies (After Implementation)

```
UI Components Architecture:
├── Model (orchestrator)
│   ├── InputComponent
│   │   └── (bubbletea TextArea)
│   ├── HistoryComponent
│   │   └── Renderer
│   │       ├── MessageFormatter
│   │       └── ToolFormatter
│   ├── ToolStatusComponent
│   │   └── Renderer
│   └── HelpComponent
│
├── ThemeManager
│   └── (5+ Theme instances)
│
└── PluginManager
    └── (User plugins)

Rendering Data Flow:
UI Model → Components → Renderers → Formatters → lipgloss → Text
```

---

## Technology Stack (Updated)

| Layer | Technology | Status |
|-------|-----------|--------|
| **Language** | Go 1.26+ | ✓ Unchanged |
| **TUI Framework** | BubbleTea v2 | ✓ Maintained |
| **Styling** | Lipgloss v2 | ✓ Maintained |
| **Rendering** | Custom ScrollingRenderer | ✨ New |
| **Themes** | Custom system (YAML) | ✨ New |
| **Plugins** | Custom framework | ✨ New |
| **LLM** | Catwalk (multi-provider) | ✓ Unchanged |
| **Database** | SQLite3 | ✓ Unchanged |
| **Removed** | Ultraviolet | ✗ Removed |

---

## Configuration Files (Updated)

### go.mod Changes
```diff
- github.com/charmbracelet/ultraviolet v0.0.0-...
+ No ultraviolet dependency
```

### New Config Locations
```
~/.ghost/config.yaml         # Main config (themes subsection added)
~/.ghost/themes/             # User theme directory (auto-created)
~/.ghost/plugins/            # User plugin directory (auto-created)
```

### Example Config
```yaml
app:
  theme: "branded"  # New: theme selection
  plugins:          # New: plugin configuration
    - name: "custom-search"
      enabled: true

providers:
  # ... existing provider config ...
```

---

## Build & Distribution

### Build Output
```
ghost-v2.0.0-darwin-amd64
ghost-v2.0.0-darwin-arm64
ghost-v2.0.0-linux-amd64
ghost-v2.0.0-linux-arm64
ghost-v2.0.0-windows-amd64.exe
```

### Installation Methods (unchanged, but better documented)
```bash
brew install ghost
apt install ghost  # Linux
choco install ghost  # Windows
```

---

## Repository Stats (Post-Implementation)

```
Total Commits Added:        ~200
Lines Changed:              +15,000 / -10,000
New Tests:                  +250
Documentation Pages:        +20
Example Projects:           +6
```

---

## Quality Metrics (Post-Implementation)

```
Code Coverage:              90%+
Linting Grade:              A+
Cyclomatic Complexity:      <15 (avg per function)
Build Time:                 <2 minutes
Test Suite Runtime:         <30 seconds
Benchmark Pass Rate:        100%
Memory Leak Detection:      ✓ Passed
```

---

## Backward Compatibility

### Maintained
- ✓ Existing sessions continue to work
- ✓ Config files remain compatible
- ✓ CLI commands (with enhancements)
- ✓ Plugin system backward compatible
- ✓ Theme system optional (has defaults)

### Migration Guide
```
Users upgrading from v1.x to v2.x:
1. No action required
2. All existing sessions work
3. New features available immediately
4. Optional: Customize theme or install plugins
```

---

## Version Map

| Version | Phase | Date | Notes |
|---------|-------|------|-------|
| v1.0.x | Current | - | Baseline |
| v1.1.0-rc1 | Phase 1 | 2026-05-01 | UI Modernization (RC) |
| v1.1.0 | Phase 1 | 2026-05-15 | First release with new UI |
| v1.2.0-rc1 | Phase 2 | 2026-06-01 | Component Decomposition (RC) |
| v1.2.0 | Phase 2 | 2026-06-15 | Refactored architecture |
| v1.3.0 | Phase 3 | 2026-07-01 | Theme system added |
| v1.4.0 | Phase 4 | 2026-07-15 | Plugin system added |
| v2.0.0 | Phase 5 | 2026-08-01 | Final release w/ full documentation |

---

## How To Use This Structure

### For Developers
1. Review `ARCHITECTURE_ANALYSIS.md` - Understand current state
2. Follow `IMPLEMENTATION_PLAN.md` - Step-by-step tasks
3. Reference `internal/ui/components/README.md` - Component guide
4. Check `docs/development/` - Development guides

### For Contributors
1. Read `docs/development/contributing.md`
2. Pick an issue or feature
3. Follow `docs/development/coding-standards.md`
4. Reference appropriate guide (theme/plugin/component)
5. Submit PR with tests & docs

### For Users
1. Read `docs/features/` guides
2. Customize theme from `examples/themes/`
3. Install plugins from community registry
4. Refer to `docs/tutorials/` for advanced usage

---

## Success Indicators

When the repository reaches this structure:

✅ **Organization:** Everything in logical, focused modules  
✅ **Testability:** >90% coverage, easy to add tests  
✅ **Maintainability:** Clear separation of concerns  
✅ **Extensibility:** Theme & plugin systems mature  
✅ **Documentation:** Complete guides for all features  
✅ **Performance:** Fast, lean, memory-efficient  
✅ **Quality:** Zero warnings, A+ linting score  

---

**Last Updated:** 2026-04-01  
**Status:** Final Structure (Post-Implementation)
