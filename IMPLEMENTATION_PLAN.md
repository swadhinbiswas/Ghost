# Ghost CLI - Complete Implementation Plan
## Perfect, Properly Organized Phase-by-Phase Execution Guide

**Version:** 1.0  
**Status:** Ready for Execution  
**Timeline:** 12 Weeks  
**Last Updated:** 2026-04-01

---

## Table of Contents

1. [Overview & Principles](#overview--principles)
2. [Phase 1: UI Modernization (Weeks 1-4)](#phase-1-ui-modernization-weeks-1-4)
3. [Phase 2: Component Decomposition (Weeks 5-7)](#phase-2-component-decomposition-weeks-5-7)
4. [Phase 3: Theme System (Weeks 8-9)](#phase-3-theme-system-weeks-8-9)
5. [Phase 4: Plugin Architecture (Week 10)](#phase-4-plugin-architecture-week-10)
6. [Phase 5: Documentation & Release (Weeks 11-12)](#phase-5-documentation--release-weeks-11-12)
7. [Testing Strategy](#testing-strategy)
8. [Deployment & Rollback](#deployment--rollback)

---

## Overview & Principles

### Core Principles

1. **Backward Compatibility:** Never break existing user workflows
2. **Incremental Delivery:** Each phase delivers value independently
3. **Test-Driven:** Write tests before implementation where possible
4. **Documentation-First:** Document decisions and patterns
5. **User-Centric:** Every change improves user experience
6. **Performance:** No render time degradation

### Quality Gates

Each phase must pass:
- ✅ All tests passing (unit + integration)
- ✅ No compiler warnings
- ✅ Code review from 2+ maintainers
- ✅ Performance benchmarks stable or improved
- ✅ Documentation up-to-date

### Metrics to Track

```
Phase Completion:
  - Lines of code changed
  - Test coverage increase
  - Performance metrics (render time, memory)
  - Bugs found & fixed
  - Time spent vs. estimated

Code Health:
  - Cyclomatic complexity
  - Test coverage percentage
  - Linting score
  - Documentation coverage
```

---

# Phase 1: UI Modernization (Weeks 1-4)

## Objective
Replace Ultraviolet fullscreen layout with linear scrolling BubbleTea output. Enable natural terminal scrolling, cleaner tool visualization.

## Expected Outcomes
- ✅ Ultraviolet completely removed from codebase
- ✅ Linear output renderer working for all message types
- ✅ Tool visualization reduced to single-line format
- ✅ All existing functionality preserved
- ✅ Performance maintained or improved

---

## Week 1: Foundation & Architecture

### Task 1.1: Create Feature Branch & Setup
**Owner:** Lead Architect  
**Duration:** 1 day  
**Deliverables:**
```bash
# Create feature branch
git checkout -b refactor/ui-modernization
git push -u origin refactor/ui-modernization

# Create changelog entry
echo "Phase 1: UI Modernization - Replace Ultraviolet with linear scrolling" >> CHANGELOG_PHASE1.md
```

### Task 1.2: Analyze Ultraviolet Usage
**Owner:** Architecture Team  
**Duration:** 2 days  
**Process:**
```bash
# Find all ultraviolet imports and usages
grep -r "uv\." internal/ui/ | tee ultraviolet_usage.txt
grep -r "ultraviolet" . --include="*.go" | tee ultraviolet_imports.txt
grep -r "Draw(" internal/ui/ | tee draw_methods.txt

# Count impact
wc -l ultraviolet_usage.txt ultraviolet_imports.txt draw_methods.txt
```

**Deliverables:**
- `docs/phase1/ultraviolet_audit.md` - Complete audit of UV usage
- `docs/phase1/removal_plan.md` - Detailed removal strategy
- Import dependency map

**Key Questions to Answer:**
- [ ] Which functions depend on `uv.Screen`?
- [ ] Which rendering logic uses `uv.Rectangle`?
- [ ] What layout information is stored in ultraviolet structures?
- [ ] How many custom Draw methods exist?

### Task 1.3: Create ScrollingRenderer Prototype
**Owner:** Core UI Developer  
**Duration:** 3 days  
**File:** `internal/ui/renderer/scrolling.go`

```go
package renderer

import (
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// ScrollingRenderer handles linear terminal output rendering
type ScrollingRenderer struct {
	mu                sync.RWMutex
	styles            *styles.Styles
	chatHistory       []RenderedMessage
	toolSpinners      map[string]ToolStatus
	inputPrompt       string
	maxHistoryLines   int
	terminalWidth     int
	enableColors      bool
}

// RenderedMessage represents a pre-rendered message
type RenderedMessage struct {
	ID        string
	Content   string
	Role      message.Role
	Height    int
	Timestamp int64
}

// ToolStatus tracks tool execution state
type ToolStatus struct {
	Name      string
	Status    string // "pending", "running", "success", "error"
	Duration  string
	Message   string
}

// NewScrollingRenderer creates a new scrolling renderer
func NewScrollingRenderer(sty *styles.Styles) *ScrollingRenderer {
	return &ScrollingRenderer{
		styles:       sty,
		chatHistory:  make([]RenderedMessage, 0, 100),
		toolSpinners: make(map[string]ToolStatus),
		maxHistoryLines: 1000,
		enableColors: true,
	}
}

// AddMessage adds a rendered message to history
func (sr *ScrollingRenderer) AddMessage(msg RenderedMessage) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.chatHistory = append(sr.chatHistory, msg)
}

// UpdateToolStatus updates tool execution status
func (sr *ScrollingRenderer) UpdateToolStatus(toolName string, status ToolStatus) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.toolSpinners[toolName] = status
}

// SetInputPrompt sets the active input prompt area
func (sr *ScrollingRenderer) SetInputPrompt(prompt string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.inputPrompt = prompt
}

// SetTerminalWidth updates terminal width for wrapping
func (sr *ScrollingRenderer) SetTerminalWidth(width int) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.terminalWidth = width
}

// Render generates the complete terminal output
func (sr *ScrollingRenderer) Render() string {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	var buf strings.Builder

	// Render chat history (with history trimming if too large)
	buf.WriteString(sr.renderChatHistory())

	// Render active tool statuses
	if len(sr.toolSpinners) > 0 {
		buf.WriteString("\n\n")
		buf.WriteString(sr.renderToolStatus())
	}

	// Render input prompt area
	if sr.inputPrompt != "" {
		buf.WriteString("\n\n")
		buf.WriteString(sr.inputPrompt)
	}

	return buf.String()
}

// renderChatHistory generates formatted chat history
func (sr *ScrollingRenderer) renderChatHistory() string {
	var buf strings.Builder
	for i, msg := range sr.chatHistory {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(sr.formatMessage(msg))
	}
	return buf.String()
}

// renderToolStatus generates tool execution status display
func (sr *ScrollingRenderer) renderToolStatus() string {
	var buf strings.Builder
	buf.WriteString("Tools Running:\n")
	for _, tool := range sr.toolSpinners {
		buf.WriteString(sr.formatToolStatus(tool))
		buf.WriteString("\n")
	}
	return buf.String()
}

// formatMessage formats a single message for display
func (sr *ScrollingRenderer) formatMessage(msg RenderedMessage) string {
	prefix := "You"
	if msg.Role == message.RoleAssistant {
		prefix = "Ghost"
	}
	return lipgloss.NewStyle().
		Bold(true).
		Render(prefix+": ") + msg.Content
}

// formatToolStatus formats tool status for display
func (sr *ScrollingRenderer) formatToolStatus(tool ToolStatus) string {
	icon := "●" // pending
	if tool.Status == "success" {
		icon = "✓"
	} else if tool.Status == "error" {
		icon = "✗"
	}

	result := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render(icon) + " " + tool.Name

	if tool.Duration != "" {
		result += " (" + tool.Duration + ")"
	}
	if tool.Message != "" {
		result += " - " + tool.Message
	}

	return result
}

// GetHeight returns the rendered height in lines
func (sr *ScrollingRenderer) GetHeight() int {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	height := 0
	for _, msg := range sr.chatHistory {
		height += msg.Height
	}
	height += len(sr.toolSpinners) // Tool status lines
	height += 2                      // Vertical spacing
	return height
}

// Clear resets all rendered content
func (sr *ScrollingRenderer) Clear() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.chatHistory = make([]RenderedMessage, 0)
	sr.toolSpinners = make(map[string]ToolStatus)
}

// TrimHistory removes oldest messages if over limit
func (sr *ScrollingRenderer) TrimHistory() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	totalLines := 0
	for _, msg := range sr.chatHistory {
		totalLines += msg.Height
	}

	if totalLines > sr.maxHistoryLines {
		// Remove oldest messages until under limit
		for totalLines > sr.maxHistoryLines && len(sr.chatHistory) > 0 {
			totalLines -= sr.chatHistory[0].Height
			sr.chatHistory = sr.chatHistory[1:]
		}
	}
}
```

**Deliverables:**
- `internal/ui/renderer/scrolling.go` - Core renderer
- `internal/ui/renderer/scrolling_test.go` - Unit tests
- Documentation: renderer API contract

### Task 1.4: Create Message Formatter
**Owner:** UI Developer  
**Duration:** 2 days  
**File:** `internal/ui/renderer/message_formatter.go`

```go
package renderer

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// MessageFormatter handles message rendering
type MessageFormatter struct {
	styles *styles.Styles
	width  int
}

// NewMessageFormatter creates a formatter
func NewMessageFormatter(sty *styles.Styles, width int) *MessageFormatter {
	return &MessageFormatter{
		styles: sty,
		width:  width,
	}
}

// FormatUserMessage formats a user message
func (mf *MessageFormatter) FormatUserMessage(msg *message.Message) string {
	content := msg.Content().Text
	wrapped := lipgloss.NewStyle().
		Width(mf.width - 5).
		Render(content)

	return mf.styles.Chat.Message.UserFocused.Render("You:\n" + wrapped)
}

// FormatAssistantMessage formats an assistant message
func (mf *MessageFormatter) FormatAssistantMessage(msg *message.Message) string {
	content := msg.Content().Text
	wrapped := lipgloss.NewStyle().
		Width(mf.width - 5).
		Render(content)

	return mf.styles.Chat.Message.AssistantFocused.Render("Ghost:\n" + wrapped)
}

// FormatToolCall formats a tool call result
func (mf *MessageFormatter) FormatToolCall(name string, success bool, result string) string {
	icon := "✗"
	style := mf.styles.ToolCallError
	if success {
		icon = "✓"
		style = mf.styles.ToolCallSuccess
	}

	return style.Render("[" + icon + "] " + name + "\n" + result)
}

// SetWidth updates the formatting width
func (mf *MessageFormatter) SetWidth(width int) {
	mf.width = width
}
```

**Deliverables:**
- `internal/ui/renderer/message_formatter.go`
- `internal/ui/renderer/message_formatter_test.go`

---

## Week 2: Core Rendering Implementation

### Task 2.1: Integrate ScrollingRenderer into UI Model
**Owner:** Core UI Developer  
**Duration:** 2 days  

**Changes to `internal/ui/model/ui.go`:**

1. Replace ultraviolet imports:
```go
// REMOVE:
// uv "github.com/charmbracelet/ultraviolet"

// ADD:
"github.com/swadhinbiswas/ghost/internal/ui/renderer"
```

2. Modify View() method:
```go
func (m *UI) View() tea.View {
	var v tea.View
	v.AltScreen = false  // KEY CHANGE: Disable fullscreen
	v.MouseMode = tea.MouseModeCellMotion
	v.ReportFocus = m.caps.ReportFocusEvents
	v.WindowTitle = "Ghost " + home.Short(m.com.Store().WorkingDir())

	// Use new renderer instead of ultraviolet canvas
	content := m.renderLinearLayout()
	v.Content = content

	return v
}

// New method using ScrollingRenderer
func (m *UI) renderLinearLayout() string {
	m.scrollRenderer.SetTerminalWidth(m.width)
	return m.scrollRenderer.Render()
}
```

3. Remove Draw() method entirely:
```go
// DELETE THIS METHOD:
// func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor
```

**Deliverables:**
- Modified `internal/ui/model/ui.go`
- Updated View() method implementation
- Removed 500+ lines of ultraviolet code

### Task 2.2: Migrate Chat Message Rendering
**Owner:** UI Developer  
**Duration:** 3 days  

**Update `internal/ui/chat/messages.go`:**
- Refactor message rendering to return strings instead of ultraviolet objects
- Remove ultraviolet layout dependencies
- Ensure markdown rendering works in linear mode

**Deliverables:**
- Updated chat message rendering functions
- All message types tested (user, assistant, thinking, tool calls)

### Task 2.3: Update Tool Visualization
**Owner:** UI Developer  
**Duration:** 2 days  

**Refactor `internal/ui/chat/tools.go`:**

```go
// New minimal tool rendering
func RenderToolPending(toolName string) string {
	return "⋯ " + toolName + "...\n"
}

func RenderToolSuccess(toolName string, duration time.Duration) string {
	return "✓ " + toolName + " (" + duration.String() + ")\n"
}

func RenderToolError(toolName string, err error) string {
	return "✗ " + toolName + " - " + err.Error() + "\n"
}

// Replaces large card-based rendering
```

**Deliverables:**
- Minimal tool rendering functions
- No more large cards; single-line status updates
- Tests for all tool states

---

## Week 3: Testing & Validation

### Task 3.1: Comprehensive Testing
**Owner:** QA + Developer  
**Duration:** 3 days  

**Test Scenarios:**

```go
// Test 1: Basic message rendering
func TestScrollingRenderer_BasicMessages(t *testing.T) {
	sr := renderer.NewScrollingRenderer(testStyles)
	sr.AddMessage(renderer.RenderedMessage{
		ID:      "msg1",
		Content: "Hello World",
		Role:    message.RoleUser,
		Height:  1,
	})
	output := sr.Render()
	assert.Contains(t, output, "You: Hello World")
}

// Test 2: Tool status updates
func TestScrollingRenderer_ToolStatus(t *testing.T) {
	sr := renderer.NewScrollingRenderer(testStyles)
	sr.UpdateToolStatus("bash", renderer.ToolStatus{
		Name:   "bash",
		Status: "success",
		Duration: "150ms",
	})
	output := sr.Render()
	assert.Contains(t, output, "✓ bash")
}

// Test 3: Long message wrapping
func TestScrollingRenderer_LongMessages(t *testing.T) {
	sr := renderer.NewScrollingRenderer(testStyles)
	sr.SetTerminalWidth(80)
	longMsg := strings.Repeat("a", 200)
	sr.AddMessage(renderer.RenderedMessage{
		Content: longMsg,
		Role:    message.RoleAssistant,
		Height:  3,
	})
	output := sr.Render()
	lines := strings.Split(output, "\n")
	assert.True(t, len(lines) >= 3)
}

// Test 4: History trimming
func TestScrollingRenderer_HistoryTrimming(t *testing.T) {
	sr := renderer.NewScrollingRenderer(testStyles)
	sr.maxHistoryLines = 10
	// Add messages that exceed limit
	for i := 0; i < 20; i++ {
		sr.AddMessage(renderer.RenderedMessage{
			Content: "msg" + strconv.Itoa(i),
			Height:  1,
		})
	}
	sr.TrimHistory()
	assert.LessOrEqual(t, sr.GetHeight(), 10)
}

// Test 5: Linear output (no ultraviolet canvas)
func TestScrollingRenderer_NoCanvasOutput(t *testing.T) {
	sr := renderer.NewScrollingRenderer(testStyles)
	sr.AddMessage(renderer.RenderedMessage{
		Content: "Test",
		Height:  1,
	})
	output := sr.Render()
	// Should be plain string, no ANSI canvas codes
	assert.NotContains(t, output, "\x1b[")  // No ESC codes for canvas
}
```

**Deliverables:**
- `internal/ui/renderer/scrolling_test.go` - Full test suite
- `internal/ui/renderer/message_formatter_test.go`
- Test coverage report (target: >90%)

### Task 3.2: Performance Benchmarking
**Owner:** Performance Engineer  
**Duration:** 2 days  

**Benchmark Suite:**

```go
func BenchmarkScrollingRenderer_Render100Messages(b *testing.B) {
	sr := renderer.NewScrollingRenderer(testStyles)
	for i := 0; i < 100; i++ {
		sr.AddMessage(renderer.RenderedMessage{
			Content: "This is message " + strconv.Itoa(i),
			Height:  1,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sr.Render()
	}
}

func BenchmarkScrollingRenderer_UpdateToolStatus(b *testing.B) {
	sr := renderer.NewScrollingRenderer(testStyles)
	tools := []string{"bash", "edit", "read", "search"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tool := range tools {
			sr.UpdateToolStatus(tool, renderer.ToolStatus{
				Name:   tool,
				Status: "running",
			})
		}
	}
}
```

**Success Criteria:**
- Render 100 messages: <50ms
- Tool status update: <1ms
- Memory usage: <100MB for 1000 messages
- No memory leaks over 1 hour

**Deliverables:**
- Benchmark results file
- Performance comparison (old vs. new)
- Optimization notes

### Task 3.3: Integration Testing
**Owner:** QA  
**Duration:** 2 days  

**Test Scenarios:**
1. Launch Ghost in interactive mode
2. Send multiple messages (mix of short/long)
3. Run multiple tools simultaneously
4. Scroll terminal history naturally
5. Verify no visual artifacts or flicker
6. Test with different terminal sizes (80x24, 120x40, etc.)
7. Test with tmux/screen multiplexers

**Deliverables:**
- Manual test checklist (PASSED)
- Screenshot comparisons (before/after)
- Bug report template for regressions

---

## Week 4: Polish & Cleanup

### Task 4.1: Remove Ultraviolet Dependencies
**Owner:** Build Engineer  
**Duration:** 1 day  

```bash
# 1. Remove from go.mod
go mod tidy
# Remove: github.com/charmbracelet/ultraviolet

# 2. Remove upstream from Makefile
grep -r "ultraviolet" . --include="Makefile" && sed -i '/ultraviolet/d' Makefile

# 3. Verify no remaining imports
grep -r "ultraviolet\|uv\." internal/ && echo "FAIL: Still has UV imports" || echo "PASS: No UV imports"

# 4. Build check
go build -v ./...
```

**Deliverables:**
- Clean `go.mod` (no ultraviolet)
- Build verification passing
- Dependency graph checked

### Task 4.2: Code Cleanup & Optimization
**Owner:** Developer  
**Duration:** 1 day  

- Remove dead code from `ui.go` (old layout system)
- Simplify state management
- Add inline documentation
- Run `gofmt` and `golangci-lint`

**Deliverables:**
- Cleaned codebase
- Linting report (0 warnings)

### Task 4.3: Update Documentation
**Owner:** Technical Writer  
**Duration:** 1 day  

**Files to Update:**
- `README.md` - Update UI description
- `CONTRIBUTING.md` - Add UI rendering guide
- `internal/ui/README.md` - New component documentation
- `docs/ui-architecture.md` - Updated architecture

**Deliverables:**
- Updated documentation
- UI architecture guide
- Contributing guidelines

### Task 4.4: Prepare Release Notes
**Owner:** Release Manager  
**Duration:** 0.5 day  

**Content:**
```markdown
## Phase 1: UI Modernization Release

### What Changed
- ✅ Replaced Ultraviolet fullscreen layout with linear scrolling output
- ✅ Simplified tool visualization (cards → single-line status)
- ✅ Disabled AltScreen mode for natural terminal scrolling
- ✅ Improved rendering performance by 30%

### Breaking Changes
- None (fully backward compatible)

### Performance Improvements
- Render time for 100 messages: 75ms → 35ms (53% faster)
- Memory usage: Reduced by 20%
- Tool status updates: <1ms

### Testing
- ✅ 120+ unit tests
- ✅ 15 integration tests
- ✅ Performance benchmarks passed

### Known Issues
- None

### Migration Guide
No action required. Existing sessions and settings continue to work.
```

**Deliverables:**
- Release notes file
- Migration guide (if needed)
- Changelog entry

### Task 4.5: Code Review & Merge
**Owner:** Maintainers  
**Duration:** 1 day  

**Review Checklist:**
- [ ] All tests passing (100% green)
- [ ] Code coverage >90%
- [ ] No performance regressions
- [ ] Documentation complete
- [ ] No linting warnings
- [ ] Backward compatible
- [ ] Commit messages clear

**Merge Process:**
```bash
# 1. Rebase on main
git rebase main refactor/ui-modernization

# 2. Squash commits (optional, for cleanliness)
git rebase -i main

# 3. Final tests
go test ./...

# 4. Merge to main
git checkout main
git merge --no-ff refactor/ui-modernization

# 5. Push
git push origin main

# 6. Tag release
git tag -a v1.0-phase1 -m "Phase 1: UI Modernization"
git push origin v1.0-phase1
```

**Deliverables:**
- Merged to main branch
- Tagged release created

---

## Phase 1 Success Criteria

### Code Metrics
- [ ] 2200+ lines removed (ultraviolet + old layout)
- [ ] 800+ lines added (new renderer)
- [ ] Net reduction: 1400+ lines of UI code
- [ ] Cyclomatic complexity reduced by 25%

### Test Coverage
- [ ] 120+ unit tests written
- [ ] 15+ integration tests added
- [ ] Snapshot tests for message rendering
- [ ] Overall coverage: 90%+

### Performance
- [ ] Render 100 messages: <50ms (baseline met)
- [ ] Tool status updates: <1ms
- [ ] Memory usage: Stable or reduced
- [ ] No memory leaks detected

### User Experience
- [ ] Chat naturally scrolls in terminal
- [ ] Tool status displays minimally
- [ ] No visual artifacts or flicker
- [ ] Works with tmux/screen/vim

### Deliverables Checklist
- [ ] `internal/ui/renderer/scrolling.go` - Core renderer
- [ ] `internal/ui/renderer/message_formatter.go` - Message formatting
- [ ] `internal/ui/renderer/*_test.go` - Complete test suite
- [ ] Updated `internal/ui/model/ui.go` - Integrated renderer
- [ ] `docs/phase1/` - Documentation folder
- [ ] Release notes & changelog
- [ ] All tests passing
- [ ] Code merged to main

---

---

# Phase 2: Component Decomposition (Weeks 5-7)

## Objective
Break monolithic UI model into focused, testable components. Improve maintainability and enable component reuse.

## Expected Outcomes
- ✅ UI model reduced from 2200 → <500 lines
- ✅ 6+ focused component modules created
- ✅ All components independently testable
- ✅ Clear component interfaces defined
- ✅ Test coverage increased to >90%

---

## Week 5: Component Design & Architecture

### Task 5.1: Define Component Interfaces
**Owner:** Architecture Team  
**Duration:** 2 days  

**File:** `internal/ui/components/interfaces.go`

```go
package components

import (
	tea "charm.land/bubbletea/v2"
	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// Component is the base interface for all UI components
type Component interface {
	// Init initializes the component
	Init(ctx context.Context) tea.Cmd
	
	// Update handles messages
	Update(msg tea.Msg) (Component, tea.Cmd)
	
	// Render outputs the component view
	Render(width, height int) string
	
	// SetWidth updates component width
	SetWidth(width int)
	
	// SetHeight updates component height
	SetHeight(height int)
	
	// Focus sets keyboard focus
	Focus()
	
	// Blur removes keyboard focus
	Blur()
	
	// IsFocused returns focus state
	IsFocused() bool
}

// InputComponent interface for text input
type InputComponent interface {
	Component
	Value() string
	SetValue(string)
	Placeholder() string
	SetPlaceholder(string)
	InsertText(string)
	DeleteChar()
	MoveCursor(direction string)
}

// HistoryComponent interface for chat history
type HistoryComponent interface {
	Component
	AddMessage(msg *message.Message)
	GetMessages() []*message.Message
	ScrollUp(lines int)
	ScrollDown(lines int)
	Clear()
}

// ToolStatusComponent interface for tool execution
type ToolStatusComponent interface {
	Component
	UpdateTool(name string, status string, message string)
	RemoveTool(name string)
	GetToolStatus(name string) string
}

// HelpComponent interface for help panel
type HelpComponent interface {
	Component
	UpdateBindings(bindings map[string]string)
	GetBindings() map[string]string
}
```

**Deliverables:**
- `internal/ui/components/interfaces.go` - Component contract
- Architecture decision document

### Task 5.2: Create Input Component
**Owner:** Developer  
**Duration:** 2 days  

**File:** `internal/ui/components/input.go`

```go
package components

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// InputComponent wraps and manages text input
type InputComponent struct {
	textarea textarea.Model
	focused  bool
	width    int
	height   int
	styles   *styles.Styles
}

// NewInputComponent creates an input component
func NewInputComponent(sty *styles.Styles) *InputComponent {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.SetHeight(3)
	ta.SetWidth(80)

	return &InputComponent{
		textarea: ta,
		focused:  false,
		width:    80,
		height:   3,
		styles:   sty,
	}
}

// Init initializes the component
func (ic *InputComponent) Init(ctx context.Context) tea.Cmd {
	return nil
}

// Update handles input messages
func (ic *InputComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	ic.textarea, cmd = ic.textarea.Update(msg)
	return ic, cmd
}

// Render outputs the input component
func (ic *InputComponent) Render(width, height int) string {
	ic.SetWidth(width)
	ic.SetHeight(height)
	return ic.textarea.View()
}

// Value returns current input text
func (ic *InputComponent) Value() string {
	return ic.textarea.Value()
}

// SetValue sets input text
func (ic *InputComponent) SetValue(text string) {
	ic.textarea.SetValue(text)
}

// Placeholder returns current placeholder
func (ic *InputComponent) Placeholder() string {
	return ic.textarea.Placeholder
}

// SetPlaceholder sets placeholder text
func (ic *InputComponent) SetPlaceholder(text string) {
	ic.textarea.Placeholder = text
}

// InsertText inserts text at cursor
func (ic *InputComponent) InsertText(text string) {
	ic.textarea.InsertString(text)
}

// DeleteChar deletes character at cursor
func (ic *InputComponent) DeleteChar() {
	ic.textarea.DeleteBeforeCursor()
}

// MoveCursor moves cursor (up/down/left/right)
func (ic *InputComponent) MoveCursor(direction string) {
	switch direction {
	case "up":
		ic.textarea.CursorUp()
	case "down":
		ic.textarea.CursorDown()
	case "left":
		ic.textarea.CursorLeft()
	case "right":
		ic.textarea.CursorRight()
	}
}

// SetWidth updates width
func (ic *InputComponent) SetWidth(width int) {
	ic.width = width
	ic.textarea.SetWidth(width - 4)
}

// SetHeight updates height
func (ic *InputComponent) SetHeight(height int) {
	ic.height = height
	ic.textarea.SetHeight(height)
}

// Focus sets focus
func (ic *InputComponent) Focus() {
	ic.focused = true
	ic.textarea.Focus()
}

// Blur removes focus
func (ic *InputComponent) Blur() {
	ic.focused = false
	ic.textarea.Blur()
}

// IsFocused returns focus state
func (ic *InputComponent) IsFocused() bool {
	return ic.focused
}
```

**Deliverables:**
- `internal/ui/components/input.go` - Input component
- `internal/ui/components/input_test.go` - Unit tests

### Task 5.3: Create History Component
**Owner:** Developer  
**Duration:** 2 days  

**File:** `internal/ui/components/history.go`

```go
package components

import (
	"context"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/swadhinbiswas/ghost/internal/message"
	"github.com/swadhinbiswas/ghost/internal/ui/renderer"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// HistoryComponent manages chat history display
type HistoryComponent struct {
	mu         sync.RWMutex
	messages   []*message.Message
	renderer   *renderer.ScrollingRenderer
	scrollPos  int
	width      int
	height     int
	styles     *styles.Styles
	formatter  *renderer.MessageFormatter
}

// NewHistoryComponent creates a history component
func NewHistoryComponent(sty *styles.Styles) *HistoryComponent {
	return &HistoryComponent{
		messages:  make([]*message.Message, 0, 100),
		renderer:  renderer.NewScrollingRenderer(sty),
		scrollPos: 0,
		width:     80,
		height:    20,
		styles:    sty,
		formatter: renderer.NewMessageFormatter(sty, 80),
	}
}

// Init initializes the component
func (hc *HistoryComponent) Init(ctx context.Context) tea.Cmd {
	return nil
}

// Update handles messages
func (hc *HistoryComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	// Handle scroll messages, etc.
	return hc, nil
}

// Render outputs chat history
func (hc *HistoryComponent) Render(width, height int) string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	hc.SetWidth(width)
	hc.SetHeight(height)

	var buf strings.Builder
	for _, msg := range hc.messages {
		buf.WriteString(hc.renderMessage(msg))
		buf.WriteString("\n\n")
	}
	return buf.String()
}

// renderMessage renders a single message
func (hc *HistoryComponent) renderMessage(msg *message.Message) string {
	if msg.Role == message.RoleUser {
		return hc.formatter.FormatUserMessage(msg)
	}
	return hc.formatter.FormatAssistantMessage(msg)
}

// AddMessage adds a message to history
func (hc *HistoryComponent) AddMessage(msg *message.Message) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.messages = append(hc.messages, msg)
}

// GetMessages returns all messages
func (hc *HistoryComponent) GetMessages() []*message.Message {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return append([]*message.Message{}, hc.messages...)
}

// ScrollUp scrolls up in history
func (hc *HistoryComponent) ScrollUp(lines int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.scrollPos += lines
}

// ScrollDown scrolls down in history
func (hc *HistoryComponent) ScrollDown(lines int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	if hc.scrollPos >= lines {
		hc.scrollPos -= lines
	}
}

// Clear clears all messages
func (hc *HistoryComponent) Clear() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.messages = make([]*message.Message, 0)
	hc.scrollPos = 0
}

// SetWidth updates width
func (hc *HistoryComponent) SetWidth(width int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.width = width
	hc.formatter.SetWidth(width)
	hc.renderer.SetTerminalWidth(width)
}

// SetHeight updates height
func (hc *HistoryComponent) SetHeight(height int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.height = height
}

// Focus sets focus (not typically focused)
func (hc *HistoryComponent) Focus() {}

// Blur removes focus
func (hc *HistoryComponent) Blur() {}

// IsFocused returns focus state (always false for history)
func (hc *HistoryComponent) IsFocused() bool {
	return false
}
```

**Deliverables:**
- `internal/ui/components/history.go` - History component
- `internal/ui/components/history_test.go` - Unit tests

### Task 5.4: Create Tool Status Component
**Owner:** Developer  
**Duration:** 1.5 days  

**File:** `internal/ui/components/tool_status.go`

```go
package components

import (
	"context"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// ToolStatus represents tool execution state
type ToolStatus struct {
	Name    string
	Status  string // "pending", "running", "success", "error"
	Message string
	Time    string
}

// ToolStatusComponent manages tool display
type ToolStatusComponent struct {
	mu    sync.RWMutex
	tools map[string]ToolStatus
	width int
	style *styles.Styles
}

// NewToolStatusComponent creates component
func NewToolStatusComponent(sty *styles.Styles) *ToolStatusComponent {
	return &ToolStatusComponent{
		tools: make(map[string]ToolStatus),
		width: 80,
		style: sty,
	}
}

// Init initializes
func (tsc *ToolStatusComponent) Init(ctx context.Context) tea.Cmd {
	return nil
}

// Update handles messages
func (tsc *ToolStatusComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return tsc, nil
}

// Render outputs tool status
func (tsc *ToolStatusComponent) Render(width, height int) string {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()

	if len(tsc.tools) == 0 {
		return ""
	}

	var buf strings.Builder
	buf.WriteString("Tools:\n")
	for _, tool := range tsc.tools {
		buf.WriteString(tsc.formatTool(tool))
		buf.WriteString("\n")
	}
	return buf.String()
}

// formatTool formats a single tool
func (tsc *ToolStatusComponent) formatTool(tool ToolStatus) string {
	icon := "⋯"
	if tool.Status == "success" {
		icon = "✓"
	} else if tool.Status == "error" {
		icon = "✗"
	}

	result := icon + " " + tool.Name
	if tool.Time != "" {
		result += " (" + tool.Time + ")"
	}
	if tool.Message != "" {
		result += " - " + tool.Message
	}
	return result
}

// UpdateTool updates tool status
func (tsc *ToolStatusComponent) UpdateTool(name, status, message string) {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	tsc.tools[name] = ToolStatus{
		Name:    name,
		Status:  status,
		Message: message,
	}
}

// RemoveTool removes tool
func (tsc *ToolStatusComponent) RemoveTool(name string) {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	delete(tsc.tools, name)
}

// GetToolStatus returns tool status
func (tsc *ToolStatusComponent) GetToolStatus(name string) string {
	tsc.mu.RLock()
	defer tsc.mu.RUnlock()
	tool, exists := tsc.tools[name]
	if !exists {
		return "unknown"
	}
	return tool.Status
}

// SetWidth updates width
func (tsc *ToolStatusComponent) SetWidth(width int) {
	tsc.mu.Lock()
	defer tsc.mu.Unlock()
	tsc.width = width
}

// SetHeight for interface compliance
func (tsc *ToolStatusComponent) SetHeight(int) {}

// Focus for interface compliance
func (tsc *ToolStatusComponent) Focus() {}

// Blur for interface compliance
func (tsc *ToolStatusComponent) Blur() {}

// IsFocused for interface compliance
func (tsc *ToolStatusComponent) IsFocused() bool {
	return false
}
```

**Deliverables:**
- `internal/ui/components/tool_status.go`
- `internal/ui/components/tool_status_test.go`

---

## Week 6: Main Component & Integration

### Task 6.1: Refactor UI Model
**Owner:** Lead Developer  
**Duration:** 3 days  

**New `internal/ui/model/ui_new.go`:**

```go
package model

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/swadhinbiswas/ghost/internal/ui/common"
	"github.com/swadhinbiswas/ghost/internal/ui/components"
)

// NewUI represents the modern, component-based UI model
type NewUI struct {
	// Components
	inputComponent      components.InputComponent
	historyComponent    components.HistoryComponent
	toolStatusComponent components.ToolStatusComponent
	helpComponent       components.HelpComponent

	// State
	com  *common.Common
	ctx  context.Context
	done chan bool

	// Dimensions
	width  int
	height int
}

// New creates a new UI instance
func NewUI(common *common.Common) *NewUI {
	return &NewUI{
		inputComponent:      components.NewInputComponent(common.Styles),
		historyComponent:    components.NewHistoryComponent(common.Styles),
		toolStatusComponent: components.NewToolStatusComponent(common.Styles),
		helpComponent:       components.NewHelpComponent(common.Styles),
		com:                 common,
	}
}

// Init initializes the UI
func (u *NewUI) Init() tea.Cmd {
	return tea.Batch(
		u.inputComponent.Init(u.ctx),
		u.historyComponent.Init(u.ctx),
		u.toolStatusComponent.Init(u.ctx),
	)
}

// Update handles messages
func (u *NewUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		u.width = msg.Width
		u.height = msg.Height
		u.inputComponent.SetWidth(u.width)
		u.historyComponent.SetWidth(u.width)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return u, tea.Quit
		}
	}

	// Update components
	var cmd tea.Cmd
	if u.inputComponent.IsFocused() {
		u.inputComponent, cmd = u.inputComponent.Update(msg).(components.InputComponent), cmd
	}

	return u, cmd
}

// View renders the UI
func (u *NewUI) View() tea.View {
	var v tea.View
	v.AltScreen = false
	v.MouseMode = tea.MouseModeCellMotion

	// Render components
	history := u.historyComponent.Render(u.width, u.height-10)
	tools := u.toolStatusComponent.Render(u.width, 3)
	input := u.inputComponent.Render(u.width, 3)

	v.Content = history + "\n" + tools + "\n" + input
	return v
}
```

**Deliverables:**
- `internal/ui/model/ui_new.go` - Refactored UI model (<500 lines)
- Migration guide for old UI

### Task 6.2: Replace Old UI Model
**Owner:** Developer  
**Duration:** 2 days  

**Migration Steps:**
1. Backup old `ui.go` to `ui_old_backup.go`
2. Update all references to use new component structure
3. Migrate state management
4. Test all functionality

**Deliverables:**
- Updated `internal/ui/model/ui.go` (refactored)
- Migration checklist completed

### Task 6.3: Complete Test Suite
**Owner:** QA + Developer  
**Duration:** 2 days  

- Input component tests (50+ tests)
- History component tests (50+ tests)
- Tool status component tests (30+ tests)
- Integration tests for component interaction

**Deliverables:**
- Complete test suite
- Coverage report (>90)
- Test documentation

---

## Week 7: Code Cleanup & Documentation

### Task 7.1: Remove Dead Code
**Owner:** Developer  
**Duration:** 1 day  

```bash
# Find dead code
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./internal/ui/...

# Remove unused functions, types, imports
# Manual review of old layout system code
```

**Deliverables:**
- Cleaned codebase
- Dead code removed

### Task 7.2: Update Documentation
**Owner:** Technical Writer  
**Duration:** 1 day  

- `docs/component-architecture.md` - Component design
- `docs/adding-components.md` - How to add new components
- `internal/ui/components/README.md` - Component directory guide
- Code comments & docstrings

**Deliverables:**
- Component documentation
- Architecture guide
- Contributing guide

### Task 7.3: Prepare Release
**Owner:** Release Manager  
**Duration:** 0.5 day  

**Deliverables:**
- Updated release notes
- Changelog entry
- Migration guide (if needed)

---

## Phase 2 Success Criteria

### Code Metrics
- [ ] Monolithic UI reduced from 2200 → <500 lines
- [ ] 6 focused components created (each <300 lines)
- [ ] Cyclomatic complexity of main model: <10
- [ ] No circular dependencies

### Test Coverage
- [ ] 200+ new tests added
- [ ] Coverage increased from 40% → >90%
- [ ] All components independently testable
- [ ] All critical paths tested

### Architecture
- [ ] Clear component interfaces defined
- [ ] Separation of concerns achieved
- [ ] No tightly coupled components
- [ ] Reusable across modules

### Deliverables Checklist
- [ ] `internal/ui/components/interfaces.go`
- [ ] `internal/ui/components/input.go` + tests
- [ ] `internal/ui/components/history.go` + tests
- [ ] `internal/ui/components/tool_status.go` + tests
- [ ] Refactored `internal/ui/model/ui.go`
- [ ] Component documentation
- [ ] All tests passing
- [ ] Code review approved

---

---

# Phase 3: Theme System (Weeks 8-9)

## Objective
Implement comprehensive, user-customizable theming system. Support built-in themes, user themes, and runtime switching.

## Expected Outcomes
- ✅ Theme abstraction layer implemented
- ✅ 5+ built-in themes available
- ✅ User-custom theme support (YAML files)
- ✅ Runtime theme switching
- ✅ Theme documentation complete

---

## Week 8: Theme Architecture

### Task 8.1: Define Theme Structure
**Owner:** Designer + Developer  
**Duration:** 1 day  

**File:** `internal/ui/theme/types.go`

```go
package theme

import "image/color"

// Theme represents a complete UI theme
type Theme struct {
	Name        string
	Description string
	Author      string
	Version     string

	// Color palette
	Colors ColorPalette

	// Component styles
	Components ComponentStyles

	// Typography
	Typography TypographyStyles
}

// ColorPalette defines all colors used
type ColorPalette struct {
	// Primary & secondary
	Primary       color.Color
	Secondary     color.Color
	Accent        color.Color

	// Status colors
	Success       color.Color
	Warning       color.Color
	Error         color.Color
	Info          color.Color

	// UI colors
	Background    color.Color
	Foreground    color.Color
	Border        color.Color
	Muted         color.Color
	HalfMuted     color.Color
	Subtle        color.Color
}

// ComponentStyles defines styling for each component
type ComponentStyles struct {
	Message MessageStyles
	Input   InputStyles
	Tools   ToolStyles
	Help    HelpStyles
}

// MessageStyles defines message styling
type MessageStyles struct {
	UserName       StyleDef
	UserContent    StyleDef
	AssistantName  StyleDef
	AssistantContent StyleDef
	Thinking       StyleDef
}

// StyleDef defines a style
type StyleDef struct {
	Color       color.Color
	Bold        bool
	Italic      bool
	Underline   bool
	Background  color.Color
}

// TypographyStyles defines typography
type TypographyStyles struct {
	FontFamily string
	LineHeight int
	LetterSpacing int
}
```

**Deliverables:**
- `internal/ui/theme/types.go` - Theme type definitions

### Task 8.2: Create Theme Manager
**Owner:** Developer  
**Duration:** 2 days  

**File:** `internal/ui/theme/manager.go`

```go
package theme

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Manager handles theme loading, switching, and persistence
type Manager struct {
	mu            sync.RWMutex
	currentTheme  *Theme
	builtinThemes map[string]*Theme
	userThemes    map[string]*Theme
	themesDir     string
}

// NewManager creates a theme manager
func NewManager(themesDir string) (*Manager, error) {
	m := &Manager{
		builtinThemes: make(map[string]*Theme),
		userThemes:    make(map[string]*Theme),
		themesDir:     themesDir,
	}

	// Load built-in themes
	if err := m.loadBuiltinThemes(); err != nil {
		return nil, err
	}

	// Load user themes
	if err := m.loadUserThemes(); err != nil {
		// Non-fatal: user themes are optional
		fmt.Println("Warning: Failed to load user themes:", err)
	}

	// Set default theme
	m.currentTheme = m.builtinThemes["dark"]

	return m, nil
}

// loadBuiltinThemes loads all built-in themes
func (m *Manager) loadBuiltinThemes() error {
	// Load from embedded assets
	builtins := map[string]*Theme{
		"dark":      NewDarkTheme(),
		"light":     NewLightTheme(),
		"branded":   NewBrandedTheme(),
		"minimal":   NewMinimalTheme(),
		"high-contrast": NewHighContrastTheme(),
	}

	m.builtinThemes = builtins
	return nil
}

// loadUserThemes loads user-defined themes from ~/.ghost/themes/
func (m *Manager) loadUserThemes() error {
	if _, err := os.Stat(m.themesDir); os.IsNotExist(err) {
		return nil // No user themes directory
	}

	files, err := ioutil.ReadDir(m.themesDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml" {
			continue
		}

		themePath := filepath.Join(m.themesDir, file.Name())
		themeName := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]

		theme, err := loadThemeFromFile(themePath)
		if err != nil {
			fmt.Printf("Failed to load theme %s: %v\n", themeName, err)
			continue
		}

		m.userThemes[themeName] = theme
	}

	return nil
}

// GetTheme returns a theme by name
func (m *Manager) GetTheme(name string) (*Theme, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check user themes first
	if theme, exists := m.userThemes[name]; exists {
		return theme, nil
	}

	// Then check built-in themes
	if theme, exists := m.builtinThemes[name]; exists {
		return theme, nil
	}

	return nil, fmt.Errorf("theme not found: %s", name)
}

// SetTheme sets the current theme
func (m *Manager) SetTheme(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	theme, err := m.GetTheme(name)
	if err != nil {
		return err
	}

	m.currentTheme = theme
	return nil
}

// GetCurrentTheme returns the active theme
func (m *Manager) GetCurrentTheme() *Theme {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTheme
}

// ListThemes returns all available themes
func (m *Manager) ListThemes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	themes := make([]string, 0)
	for name := range m.builtinThemes {
		themes = append(themes, name)
	}
	for name := range m.userThemes {
		themes = append(themes, name)
	}
	return themes
}

// SaveUserTheme saves a theme as a user theme file
func (m *Manager) SaveUserTheme(theme *Theme) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(m.themesDir, 0755); err != nil {
		return err
	}

	themePath := filepath.Join(m.themesDir, theme.Name+".yaml")
	data, err := yaml.Marshal(theme)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(themePath, data, 0644)
}

// loadThemeFromFile loads a theme from a YAML file
func loadThemeFromFile(path string) (*Theme, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	theme := &Theme{}
	if err := yaml.Unmarshal(data, theme); err != nil {
		return nil, err
	}

	return theme, nil
}
```

**Deliverables:**
- `internal/ui/theme/manager.go` - Theme manager
- `internal/ui/theme/manager_test.go` - Unit tests

### Task 8.3: Create Built-in Themes
**Owner:** Designer + Developer  
**Duration:** 1.5 days  

**File:** `internal/ui/theme/builtins.go`

```go
package theme

import (
	"image/color"

	"charm.land/x/exp/charmtone"
)

// NewDarkTheme returns the default dark theme
func NewDarkTheme() *Theme {
	return &Theme{
		Name:        "dark",
		Description: "Default dark theme",
		Author:      "Ghost",
		Version:     "1.0",
		Colors: ColorPalette{
			Primary:       charmtone.Cyan,
			Secondary:     charmtone.Magenta,
			Accent:        charmtone.Blue,
			Success:       charmtone.Green,
			Warning:       charmtone.Yellow,
			Error:         charmtone.Red,
			Info:          charmtone.Cyan,
			Background:    color.RGBA{10, 14, 39, 255},
			Foreground:    color.RGBA{224, 224, 224, 255},
			Border:        charmtone.BrightBlack,
			Muted:         color.RGBA{96, 96, 96, 255},
			HalfMuted:     color.RGBA{128, 128, 128, 255},
			Subtle:        color.RGBA{80, 80, 80, 255},
		},
	}
}

// NewLightTheme returns a light theme
func NewLightTheme() *Theme {
	return &Theme{
		Name:        "light",
		Description: "Light theme for daytime use",
		Author:      "Ghost",
		Version:     "1.0",
		Colors: ColorPalette{
			Primary:       charmtone.Blue,
			Secondary:     charmtone.Purple,
			Accent:        charmtone.Cyan,
			Success:       charmtone.Green,
			Warning:       charmtone.Yellow,
			Error:         charmtone.Red,
			Info:          charmtone.Blue,
			Background:    color.RGBA{245, 245, 245, 255},
			Foreground:    color.RGBA{31, 31, 31, 255},
			Border:        color.RGBA{200, 200, 200, 255},
			Muted:         color.RGBA{128, 128, 128, 255},
			HalfMuted:     color.RGBA{160, 160, 160, 255},
			Subtle:        color.RGBA{192, 192, 192, 255},
		},
	}
}

// NewBrandedTheme returns the custom branded theme
func NewBrandedTheme() *Theme {
	return &Theme{
		Name:        "branded",
		Description: "Ghost branded theme with custom colors",
		Author:      "Ghost",
		Version:     "1.0",
		Colors: ColorPalette{
			Primary:       color.RGBA{0, 217, 255, 255},   // Cyan
			Secondary:     color.RGBA{255, 0, 110, 255},    // Pink
			Accent:        color.RGBA{6, 255, 165, 255},    // Green
			Success:       color.RGBA{6, 255, 165, 255},    // Green
			Warning:       color.RGBA{255, 183, 3, 255},    // Orange
			Error:         color.RGBA{255, 0, 110, 255},    // Pink
			Info:          color.RGBA{0, 217, 255, 255},    // Cyan
			Background:    color.RGBA{10, 14, 39, 255},     // Very dark blue
			Foreground:    color.RGBA{224, 224, 224, 255},  // Light gray
			Border:        color.RGBA{80, 80, 120, 255},    // Dark blue-gray
			Muted:         color.RGBA{96, 96, 96, 255},     // Dark gray
			HalfMuted:     color.RGBA{128, 128, 128, 255},  // Gray
			Subtle:        color.RGBA{60, 60, 90, 255},     // Dark blue-gray
		},
	}
}

// NewMinimalTheme returns a minimalist monochrome theme
func NewMinimalTheme() *Theme {
	return &Theme{
		Name:        "minimal",
		Description: "Minimalist monochrome theme",
		Author:      "Ghost",
		Version:     "1.0",
		Colors: ColorPalette{
			Primary:       color.RGBA{200, 200, 200, 255},
			Secondary:     color.RGBA{180, 180, 180, 255},
			Accent:        color.RGBA{220, 220, 220, 255},
			Success:       color.RGBA{160, 160, 160, 255},
			Warning:       color.RGBA{180, 180, 180, 255},
			Error:         color.RGBA{140, 140, 140, 255},
			Info:          color.RGBA{190, 190, 190, 255},
			Background:    color.RGBA{30, 30, 30, 255},
			Foreground:    color.RGBA{220, 220, 220, 255},
			Border:        color.RGBA{100, 100, 100, 255},
			Muted:         color.RGBA{80, 80, 80, 255},
			HalfMuted:     color.RGBA{110, 110, 110, 255},
			Subtle:        color.RGBA{60, 60, 60, 255},
		},
	}
}

// NewHighContrastTheme returns an accessible high-contrast theme
func NewHighContrastTheme() *Theme {
	return &Theme{
		Name:        "high-contrast",
		Description: "High contrast theme for accessibility",
		Author:      "Ghost",
		Version:     "1.0",
		Colors: ColorPalette{
			Primary:       color.RGBA{255, 255, 0, 255},    // Bright yellow
			Secondary:     color.RGBA{0, 255, 255, 255},    // Bright cyan
			Accent:        color.RGBA{255, 0, 255, 255},    // Bright magenta
			Success:       color.RGBA{0, 255, 0, 255},      // Bright green
			Warning:       color.RGBA{255, 255, 0, 255},    // Bright yellow
			Error:         color.RGBA{255, 0, 0, 255},      // Bright red
			Info:          color.RGBA{0, 255, 255, 255},    // Bright cyan
			Background:    color.RGBA{0, 0, 0, 255},        // Pure black
			Foreground:    color.RGBA{255, 255, 255, 255},  // Pure white
			Border:        color.RGBA{255, 255, 255, 255},  // White
			Muted:         color.RGBA{128, 128, 128, 255},  // Gray
			HalfMuted:     color.RGBA{160, 160, 160, 255},  // Light gray
			Subtle:        color.RGBA{100, 100, 100, 255},  // Dark gray
		},
	}
}
```

**Deliverables:**
- `internal/ui/theme/builtins.go` - 5 built-in themes
- Theme design documentation

---

## Week 9: Theme Integration & Documentation

### Task 9.1: Integrate Theme Manager into UI
**Owner:** Developer  
**Duration:** 2 days  

**Updates to `internal/ui/model/ui.go`:**

```go
// Add ThemeManager to UI struct
type UI struct {
    // ... existing fields ...
    themeManager *theme.Manager
}

// Update colors dynamically based on theme
func (u *UI) updateTheme(themeName string) error {
    err := u.themeManager.SetTheme(themeName)
    if err != nil {
        return err
    }
    
    currentTheme := u.themeManager.GetCurrentTheme()
    u.updateStylesFromTheme(currentTheme)
    
    return nil
}

// Handle theme change command
func (u *UI) handleThemeCommand(args []string) {
    if len(args) < 1 {
        // List available themes
        themes := u.themeManager.ListThemes()
        // Display to user
        return
    }
    
    err := u.updateTheme(args[0])
    if err != nil {
        // Show error to user
    }
}
```

**Deliverables:**
- Integrated theme manager into UI
- Theme switching commands
- Theme persistence in config

### Task 9.2: Create Theme Documentation
**Owner:** Technical Writer  
**Duration:** 1.5 days  

**Files:**
- `docs/theming.md` - User guide for theming
- `docs/creating-themes.md` - Theme developer guide
- `examples/themes/` - Theme examples
- Theme YAML schema documentation

**Content Outline:**
- Available built-in themes
- How to create custom themes
- Theme YAML structure
- Color palette reference
- Theme validation

**Deliverables:**
- Complete theming documentation
- 3+ example theme files
- Theme creation guide

### Task 9.3: Create Theme CLI Commands
**Owner:** Developer  
**Duration:** 1 day  

**New Commands:**
```
ghost config theme list          # List available themes
ghost config theme set <name>    # Set active theme
ghost config theme preview <name> # Preview a theme
ghost config theme export <name> # Export theme as YAML
ghost config theme import <file> # Import theme from file
```

**Deliverables:**
- Theme CLI commands implemented
- Command help text
- Tests for commands

---

## Phase 3 Success Criteria

### Theming System
- [ ] Theme abstraction layer implemented
- [ ] 5 built-in themes available
- [ ] User theme loading from `~/.ghost/themes/`
- [ ] Runtime theme switching working
- [ ] Theme persistence in user config

### Built-in Themes
- [ ] Dark theme (default)
- [ ] Light theme
- [ ] Branded theme (custom Ghost colors)
- [ ] Minimal monochrome theme
- [ ] High-contrast accessibility theme

### User Experience
- [ ] Easy theme switching
- [ ] Theme preview capability
- [ ] Custom theme support
- [ ] No visual artifacts on theme change
- [ ] Themes persist across sessions

### Documentation
- [ ] Theming user guide
- [ ] Theme developer guide
- [ ] YAML schema documentation
- [ ] 3+ example themes

### Deliverables Checklist
- [ ] `internal/ui/theme/` directory with all files
- [ ] Theme manager implementation
- [ ] 5 built-in themes
- [ ] Theme CLI commands
- [ ] Documentation files
- [ ] Example themes
- [ ] All tests passing

---

---

# Phase 4: Plugin Architecture (Week 10)

## Objective
Implement plugin/extension system for custom tools and features.

## Expected Outcomes
- ✅ Plugin discovery & loading system
- ✅ Plugin interface defined
- ✅ Plugin manager implemented
- ✅ Example plugins created

**[See ARCHITECTURE_ANALYSIS.md Phase 4 for detailed specifications]**

---

# Phase 5: Documentation & Release (Weeks 11-12)

## Objective
Complete documentation, testing, and prepare for release.

## Expected Outcomes
- ✅ Complete developer documentation
- ✅ API reference documentation
- ✅ Contributor guide finalized
- ✅ Release notes prepared
- ✅ Final QA passed

**[See ARCHITECTURE_ANALYSIS.md Phase 5 for detailed specifications]**

---

---

# Testing Strategy

## Test Pyramid

```
         /\
        /  \       E2E Tests (10%)
       /────\      Integration Tests (30%)
      /      \     Unit Tests (60%)
     /________\
```

### Unit Tests (60%)
- Component rendering
- MessageFormatter functions
- ToolStatus updates
- Theme loading & switching
- Config parsing

**Target:** 90%+ coverage per module

### Integration Tests (30%)
- Component interaction
- Full rendering pipeline
- Theme switching with UI
- Plugin loading

**Target:** All critical workflows

### E2E Tests (10%)
- Full user scenarios
- Multi-tool execution
- Theme switching during session
- Long-running stability

### Performance Tests
- Render 100 messages: <50ms
- Tool status update: <1ms
- Theme switch: <100ms
- Memory stability over 1 hour

---

# Deployment & Rollback

## Deployment Process

### Pre-Deployment
- [ ] All tests passing (unit, integration, E2E)
- [ ] Code reviewed by 2+ maintainers
- [ ] Performance benchmarks passed
- [ ] Documentation complete
- [ ] Release notes prepared

### Deployment
```bash
# 1. Create release branch
git checkout -b release/v1.1.0

# 2. Update version
sed -i 's/v1.0.0/v1.1.0/g' internal/version/version.go

# 3. Update CHANGELOG
echo "## v1.1.0 - 2026-06-01" >> CHANGELOG.md

# 4. Commit and tag
git commit -am "Release v1.1.0: UI Modernization & Component Decomposition"
git tag -a v1.1.0 -m "Release v1.1.0"

# 5. Push to main
git checkout main
git merge --no-ff release/v1.1.0
git push origin main
git push origin v1.1.0

# 6. Build & publish
goreleaser release --clean
```

### Post-Deployment
- [ ] Verify binaries published
- [ ] Test on multiple OS/architectures
- [ ] Monitor error reporting
- [ ] Gather user feedback

### Rollback (if needed)
```bash
# Revert to previous version
git revert v1.1.0
git tag -a v1.1.0-revert -m "Rollback from v1.1.0"
goreleaser release --clean

# Communicate with users
# Document issue that triggered rollback
```

---

## Summary

This comprehensive implementation plan provides a complete roadmap for transforming Ghost from "good" to "perfect":

✅ **Phase 1 (Weeks 1-4):** Modern scrolling UI  
✅ **Phase 2 (Weeks 5-7):** Component architecture  
✅ **Phase 3 (Weeks 8-9):** Theme system  
✅ **Phase 4 (Week 10):** Plugin architecture  
✅ **Phase 5 (Weeks 11-12):** Documentation & release  

Each phase is independently valuable and can be merged incrementally. Follow quality gates religiously, and Ghost will emerge as a best-in-class terminal AI assistant.

---

**Document Version:** 1.0  
**Status:** Ready for Execution  
**Last Updated:** 2026-04-01
