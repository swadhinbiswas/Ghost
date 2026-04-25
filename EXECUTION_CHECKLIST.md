# Ghost UI v2.0 - Quick Start Execution Guide

## 🚀 Start Here if You're Ready to Code

This is your day-by-day reference guide. Print it. Follow it. Win. ✨

---

## WEEK 1: LAYOUT FOUNDATION

### Day 1: Planning & Branching

**Morning:**
```bash
# Start fresh branch
cd /home/swadhin/Ghost
git checkout -b refactor/ui-v2-crush-architecture
git push -u origin refactor/ui-v2-crush-architecture

# Verify current state
ls -la internal/ui/model/ui.go
wc -l internal/ui/model/ui.go  # Should be 2200+

# Document baseline
git log --oneline -5 > MIGRATION_BASELINE.txt
```

**Afternoon:**
- [ ] Read GHOST_UI_V2_CRUSH_PLAN.md completely
- [ ] Read internal/ui/model/ui.go (understand current structure)
- [ ] Identify all UI regions (chat, tools, input, help)
- [ ] Note current pixel/line heights for each region
- [ ] Create design document

**Evening:**
- [ ] Team review of design
- [ ] Get approval to proceed

---

### Day 2-3: Layout Implementation

**Task:** Create `internal/ui/layout.go` (~100 lines)

```bash
# Create files
touch internal/ui/layout.go
touch internal/ui/layout_test.go

# Copy template from GHOST_UI_V2_CRUSH_PLAN.md Phase 1, Week 1
# Update for Ghost's specific dimensions
```

**Code Tasks:**
- [ ] Define `Layout` struct with 4 rectangles (Chat, Tools, Input, Help)
- [ ] Implement `CalculateLayout(width, height int) Layout`
- [ ] Handle edge cases (small terminals, rotations)
- [ ] Implement `Layout.Valid()` checker

**Testing:**
```bash
go test ./internal/ui -run TestCalculateLayout -v

# If fails:
# - Check rectangle boundaries don't overlap
# - Verify all heights sum to terminal height
# - Test size: 80x24, 120x40, 40x10
```

**Deliverables:**
- [ ] `layout.go` ready for review
- [ ] `layout_test.go` with >80% coverage
- [ ] Benchmark: `go test -bench=. ./internal/ui`

**Git:**
```bash
git add internal/ui/layout.go internal/ui/layout_test.go
git commit -m "phase1: add layout calculation system

- Create Layout struct with Chat/Tools/Input/Help regions
- Implement CalculateLayout() for dynamic sizing
- Add rectangle boundary validation
- Comprehensive test coverage for edge cases"
```

---

### Day 4-5: UI Model Integration

**Task:** Update `internal/ui/model/ui.go` to use Layout

**What to change:**

1. **In UI struct:**
   ```go
   // ADD:
   layout Layout  // After existing fields
   ```

2. **In Init():**
   ```go
   // ADD after existing init:
   m.layout = CalculateLayout(m.width, m.height)
   ```

3. **In handleResize/WindowSize handler:**
   ```go
   // ADD after updating width/height:
   m.layout = CalculateLayout(m.width, m.height)
   ```

4. **Pass layout to Draw()** (don't call Draw yet, just prepare)

**Testing:**
```bash
go test ./internal/ui -v

# Run existing UI tests - should all pass
# No visual changes expected yet
```

**Verification:**
```bash
go build ./cmd/ghost
./ghost  # Should look identical to before

# Test resize by dragging terminal window - should adjust gracefully
```

**Git:**
```bash
git add internal/ui/model/ui.go
git commit -m "phase1: integrate layout system into UI model

- Update UI struct to store Layout instance
- Calculate layout on init and resize
- Pass layout through Draw() call chain
- All existing tests pass, no visual changes"
```

**End of Week 1 Status:**
- ✅ Location system extracted and tested
- ✅ UI model updated to use Layout
- ✅ All existing tests pass
- ✅ Ready for component extraction

---

## WEEK 2: COMPONENT STRUCTURE

### Day 6-7: Prepare Components Directory

**Morning:**
```bash
# Create components package
mkdir -p internal/ui/components
touch internal/ui/components/chat.go
touch internal/ui/components/chat_test.go
touch internal/ui/components/tools.go
touch internal/ui/components/tools_test.go
touch internal/ui/components/input.go
touch internal/ui/components/input_test.go
touch internal/ui/components/help.go
touch internal/ui/components/help_test.go
touch internal/ui/components/types.go  # Shared types
```

**Create** `internal/ui/components/types.go`:
```go
package components

import tea "charm.land/bubbletea/v2"

// Message events sent between components
type MessageAddedMsg struct {
    Message interface{}
}

type ToolUpdateMsg struct {
    Name   string
    Status string
}

type ScrollMsg struct {
    Component string
    Direction string  // "up", "down"
    Lines     int
}
```

**Day 8: ChatComponent Implementation**

Copy template from GHOST_UI_V2_CRUSH_PLAN.md Phase 2, Week 3

**Checklist:**
- [ ] NewChatComponent() works
- [ ] AddMessage() appends correctly
- [ ] ScrollUp/ScrollDown work bidirectionally
- [ ] ScrollToBottom() is called when adding messages
- [ ] VisibleMessages() returns correct slice
- [ ] Draw() renders in rectangle
- [ ] Tests all pass

**Test:**
```bash
go test ./internal/ui/components -run TestChat -v
```

---

### Day 9: ToolsComponent Implementation

**Template pattern - like ChatComponent but for tool status**

**Key methods:**
- `NewToolsComponent() *ToolsComponent`
- `UpdateTask(name string, status string)`
- `Draw(scr uv.Screen, area uv.Rectangle)`

**Each tool shows:**
- [ ] Status icon (⋯, ✓, ✗)
- [ ] Tool name
- [ ] Time elapsed
- [ ] Action (e.g., "Running bash...")

---

### Day 10-11: InputComponent & HelpComponent

**InputComponent:**
- Wraps BubbleTea's textarea
- Handles text input with cursor
- Shows placeholder

**HelpComponent:**
- Displays keybindings
- Context-aware (shows relevant keys only)
- Collapsible

**Git:**
```bash
git add internal/ui/components/
git commit -m "phase2: create isolated UI components

- ChatComponent: scrollable chat history with viewport
- ToolsComponent: real-time tool status table
- InputComponent: text input with cursor management
- HelpComponent: context-aware keybinding help
- Each <150 lines, fully tested (>85% coverage)"
```

---

## WEEK 3: MODEL REFACTORING

### Day 12-13: Update UI Model

**CRITICAL TASK:** Simplify `Draw()` from 2200 lines to 20 lines

**Current (2200 lines):**
```go
func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
    // EVERYTHING HERE (2200 lines of layout, styling, rendering)
}
```

**New (20 lines):**
```go
func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
    screen.Clear(scr)
    
    m.chat.Draw(scr, m.layout.Chat)
    m.tools.Draw(scr, m.layout.Tools)
    m.input.Draw(scr, m.layout.Input)
    m.help.Draw(scr, m.layout.Help)
    
    return m.input.CursorPosition()
}
```

**Steps:**

1. **Add components to UI struct**
   ```go
   type UI struct {
       // ... existing fields ...
       chat   *components.ChatComponent
       tools  *components.ToolsComponent
       input  *components.InputComponent
       help   *components.HelpComponent
   }
   ```

2. **Initialize in Init()**
   ```go
   func (m *UI) Init() tea.Cmd {
       m.chat = components.NewChatComponent()
       m.tools = components.NewToolsComponent()
       m.input = components.NewInputComponent()
       m.help = components.NewHelpComponent()
       // ... rest of init ...
   }
   ```

3. **Replace Draw() method**
   - Copy original Draw() to `draw_old.go` (backup)
   - Implement new 20-line Draw()
   - Delete old Draw() code

4. **Test rendering**
   ```bash
   go build ./cmd/ghost
   ./ghost
   
   # Visual test:
   # - Chat appears at top ✓
   # - Tools in middle ✓
   # - Input at bottom ✓
   # - Help bar at very bottom ✓
   ```

### Day 14-15: Update Update() Method

**Handle routing to components:**

```go
func (m *UI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c":
            return m, tea.Quit
            
        case "page-up":
            m.chat.ScrollUp(m.layout.Chat.Height() / 2)
            
        case "page-down":
            m.chat.ScrollDown(m.layout.Chat.Height() / 2)
            
        default:
            // Pass to input
            updatedInput, cmd := m.input.Update(msg)
            if cmd != nil {
                cmds = append(cmds, cmd)
            }
        }
        
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.layout = layout.CalculateLayout(m.width, m.height)
        
    case tea.Cmd:
        cmds = append(cmds, msg)
    }
    
    return m, tea.Batch(cmds...)
}
```

**Git:**
```bash
git add internal/ui/model/ui.go

# Remove old Draw() implementation
git rm internal/ui/model/ui_old.go  # if you backed it up

git commit -m "phase3: refactor UI model with component orchestration

- Reduce Draw() from 2200 lines to 20 lines
- Add component instances (chat, tools, input, help)
- Update Update() to delegate to components
- Refactor View() for clarity
- All existing behavior preserved"
```

---

## WEEK 4: TESTING & POLISH

### Day 16-17: Integration Testing

**Create** `internal/ui/integration_test.go`:

```go
package ui

import (
    "testing"
    tea "charm.land/bubbletea/v2"
)

func TestUIRenders(t *testing.T) {
    m := NewUI()
    m.width = 80
    m.height = 24
    
    v := m.View()
    if v.Content == "" {
        t.Fatal("UI returned empty content")
    }
}

func TestUIHandlesResize(t *testing.T) {
    m := NewUI()
    
    newModel, _ := m.Update(tea.WindowSizeMsg{
        Width:  120,
        Height: 40,
    })
    
    ui := newModel.(*UI)
    if ui.width != 120 || ui.height != 40 {
        t.Errorf("UI did not resize")
    }
}
```

**Run full test suite:**
```bash
go test ./... -v -cover

# Check coverage
go test ./... -cover | grep total
# Target: >85% coverage

# Profile render performance
go test -bench=. -benchmem ./internal/ui
```

### Day 18-19: Manual Testing

**Test Matrix:**

| Terminal | Behavior |
|----------|----------|
| iTerm2 (macOS) | [ ] Render, [ ] Scroll, [ ] Resize |
| GNOME Terminal (Linux) | [ ] Render, [ ] Scroll, [ ] Resize |
| Windows Terminal | [ ] Render, [ ] Scroll, [ ] Resize |
| tmux pane | [ ] Render, [ ] Scroll |
| SSH session | [ ] Render |

**Scroll testing:**
```bash
# In running ghost:
# Page Up / Page Down should scroll chat
# Large message scrolls should be smooth
# Scroll to top/bottom should work
```

**Resize testing:**
```bash
# Drag terminal edges while ghost running
# Should reflow without crashing
# No visual glitches
```

### Day 20: Performance Optimization

**Profile**:
```bash
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/ui
go tool pprof cpu.prof

# Look for hot spots, bottlenecks
```

**Targets:**
- [ ] Single render <50ms
- [ ] Memory usage <50MB
- [ ] No memory leaks

**Git:**
```bash
git add internal/ui/integration_test.go
git commit -m "phase4: comprehensive testing and optimization

- Add integration tests for UI rendering
- Test window resize handling
- Performance profiling (render time <50ms)
- Manual testing on 4+ terminal types
- All tests pass, >85% coverage
- No performance regressions"
```

---

## WEEK 5: THEME SYSTEM

### Day 21-22: Theme Structure

**Create** `internal/ui/theme/theme.go`:
- [ ] ColorScheme struct
- [ ] StyleSet struct
- [ ] Built-in themes (dark, light)
- [ ] Load/Save from file

**Create** `internal/ui/theme/themes.go`:
- [ ] DarkTheme definition
- [ ] LightTheme definition
- [ ] BrandedTheme definition

### Day 23-24: Integrate Themes

**Update components** to use theme colors:

```go
// In ChatComponent.Draw():
styledLine := theme.Current.Styles.UserMessage.Render(line)
scr.DrawString(styledLine, nil, area, y)
```

### Day 25: CLI Commands

Add commands:
```bash
ghost theme list      # Show available themes
ghost theme set dark  # Switch theme
ghost theme export    # Export current theme
```

**Git:**
```bash
git add internal/ui/theme/
git commit -m "phase5: theme system with built-in themes

- Implement ColorScheme and StyleSet system
- Create 5 built-in themes (dark, light, branded, minimal, high-contrast)
- Add theme persistence via JSON
- CLI commands for theme management
- All components respect active theme"
```

---

## WEEK 6-8: DOCUMENTATION & RELEASE

### Documentation Tasks

- [ ] Write `docs/ARCHITECTURE.md`
- [ ] Write `docs/COMPONENTS.md` (dev guide)
- [ ] Write `docs/THEMING.md` (user guide)
- [ ] Write `docs/MIGRATION.md` (from v1.x)
- [ ] Update `README.md` with new features
- [ ] Create `CHANGELOG.md`

### Release Tasks

- [ ] Final test run on all platforms
- [ ] Accessibility audit
- [ ] Security review
- [ ] Tag v2.0.0
- [ ] Push to registry

```bash
git tag -a v2.0.0 -m "Ghost UI v2.0.0: Crush-based architecture

- Refactored 2200-line UI model to 500-line component system
- Component-based rendering with layout management
- Built-in theme system with 5 themes
- Per-pane scrolling and native viewport support
- >90% test coverage and <50ms render time
- Full documentation and migration guide"

git push origin v2.0.0
```

---

## Daily Status Template

**Copy this and fill daily:**

```
## Day [X] Status

**Completed:**
- [ ] Task 1
- [ ] Task 2

**In Progress:**
- [ ] Task 3

**Blockers:**
None

**Notes:**
- Performance is good
- Tests passing

**Next Day:**
- Start Task 4
```

---

## Emergency Rollback

If critical issues:

```bash
# Revert to previous stable state
git reset --hard HEAD~1

# Or revert entire phase
git revert --no-commit HEAD~5..HEAD
git commit -m "Revert phase [X] due to [reason]"

# Or create emergency branch
git checkout -b emergency/ui-fix-critical-issue
# Fix the issue
# PR for emergency review
```

---

## Code Review Checklist

**Before committing:**

- [ ] Tests written and passing
- [ ] No regressions
- [ ] Code style consistent
- [ ] Commit message clear
- [ ] No secrets in code
- [ ] Performance checked

**Peer review:**

- [ ] Follows design
- [ ] Tests adequate
- [ ] Logic sound
- [ ] No technical debt
- [ ] Documentation updated

---

## Success Criteria (Weekly)

| Week | Metric | Target |
|------|--------|--------|
| 1 | Layout tests | 100% pass |
| 2 | Component count | 4 created |
| 3 | Draw() lines | <50 |
| 4 | Test coverage | >85% |
| 5 | Themes | 5 built-in |
| 6-8 | Release | v2.0.0 tagged |

---

## Key Files to Monitor

**Watch these files for size/complexity:**

```bash
# Should DECREASE
wc -l internal/ui/model/ui.go
# Target: 2200 → <500

# Should INCREASE (newly created)
wc -l internal/ui/components/*.go | tail -1
# Target: ~500 total

# Should INCREASE (tests)
wc -l internal/ui/*_test.go | tail -1
# Target: >1000

# Should STAY SAME
wc -l internal/ui/layout.go
# Fix: ~100 lines
```

---

## Helpful Commands

```bash
# Run tests with coverage
go test ./... -cover

# See coverage HTML
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Lint everything
golangci-lint run ./...

# Build and test
go build ./cmd/ghost && ./ghost

# Watch for changes (if using entr)
find . -name "*.go" | entr go test ./...

# Git shortcuts
git status
git diff --stat
git log --oneline refactor/ui-v2-crush-architecture -10
```

---

## Support Resources

- **Plan Details:** See `GHOST_UI_V2_CRUSH_PLAN.md`
- **Architecture:** See `CRUSH_ULTRAVIOLET_DECISION.md`
- **Ultraviolet Ref:** See `ULTRAVIOLET_REMOVAL_EXPLAINED.md`
- **Crush Source:** https://github.com/charmbracelet/crush/blob/main/internal/ui/model/ui.go

---

## You've Got This! 🚀

This plan is detailed but **100% achievable** in 8 weeks with 2 engineers.

**Key principles to remember:**

1. ✅ **Build incrementally** - Each phase is shippable
2. ✅ **Test as you go** - Don't defer testing
3. ✅ **Keep it simple** - Each component <150 lines
4. ✅ **Reference Crush** - When stuck, look at how Crush does it
5. ✅ **Communicate** - Daily standups, weekly reviews

**First commit?**

```bash
git checkout -b refactor/ui-v2-crush-architecture
echo "🚀 Starting Ghost UI v2.0 refactor - Crush-based architecture" > START.md
git add START.md
git commit -m "🚀 Start Ghost UI v2.0: Crush-based architecture"
git push -u origin refactor/ui-v2-crush-architecture
```

**Then come back to this checklist and mark off Day 1 tasks.** ✨

Let's build something beautiful. 💘
