# Ultraviolet Decision: Crush vs Ghost Analysis

## 🔴 IMPORTANT DISCOVERY

**Crush DOES use Ultraviolet.**

The official CharMBracelet reference implementation for AI-powered coding uses Ultraviolet. This changes the decision calculus for Ghost.

---

## Current Architecture Comparison

### Crush (Reference Implementation)
```go
// CharMBracelet's official approach
func (m *UI) View() tea.View {
    var v tea.View
    v.AltScreen = true              // ✓ Fullscreen mode
    v.MouseMode = tea.MouseModeCellMotion
    
    canvas := uv.NewScreenBuffer(m.width, m.height)  // ✓ Uses Ultraviolet
    v.Cursor = m.Draw(canvas, canvas.Bounds())
    
    content := strings.ReplaceAll(canvas.Render(), "\r\n", "\n")
    v.Content = content
    return v
}

func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
    // Draw components with layout rectangles
    screen.Clear(scr)
    m.chat.Draw(scr, layout.main)
    editor.Draw(scr, layout.editor)
    // ... more components
}

// BUT: Still has natural scrolling!
func (m *Chat) ScrollByAndAnimate(lines int) tea.Cmd {
    m.ScrollBy(lines)  // ✓ Viewport scrolling within the pane
    return m.RestartPausedVisibleAnimations()
}
```

### Ghost (Current State)
```go
// Current Ghost implementation (similar pattern from your analysis)
func (m *UI) View() tea.View {
    canvas := uv.NewScreenBuffer(m.width, m.height)
    m.Draw(canvas, canvas.Bounds())
    // No natural scrolling - just renders fixed content
}

func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) {
    // Complex positioning of all UI elements
    // 2200+ lines of layout logic
    // ❌ No viewport/scrolling support per component
}
```

---

## The Real Problem (Not Ultraviolet Itself)

✅ Ultraviolet is fine - it's used in Crush successfully

❌ Ghost's problem: **Incomplete implementation of Ultraviolet**

Ghost's 2200-line UI model:
- Uses Ultraviolet's canvas
- But DOESN'T implement per-component scrolling viewports
- Forces absolute positioning for everything
- No layout structure (unlike Crush's `layout.main`, `layout.editor`)

---

## Three Path Options

### Option A: FIX the Ultraviolet Usage (Recommended Given Crush Reference)
```go
// Follow Crush's actual pattern

// 1. Create proper layout rectangles
type Layout struct {
    Chat    uv.Rectangle    // Scrollable chat pane
    Tools   uv.Rectangle    // Tool status pane
    Input   uv.Rectangle    // Input area
    Help    uv.Rectangle    // Help bar
}

// 2. Each component manages its own viewport
type ChatComponent struct {
    viewport *viewport.Model  // ✓ Per-component scrolling
}

func (c *ChatComponent) Draw(scr uv.Screen, area uv.Rectangle) {
    // Draw only visible portion of chat in the rectangle
    for i, msg := range c.visibleMessages(area) {
        scr.DrawString(msg, style, area, i)
    }
}

// 3. Main UI just composes them
func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) {
    m.chat.Draw(scr, m.layout.Chat)      // Scrollable
    m.tools.Draw(scr, m.layout.Tools)    // Fixed
    m.input.Draw(scr, m.layout.Input)    // Fixed
    m.help.Draw(scr, m.layout.Help)      // Fixed
}

// Result: Crush-like architecture with proper scrolling per pane
```

**Benefits:**
- Pro: Uses proven Crush architecture
- Pro: Natural scrolling within panes (like Crush)
- Pro: Clean layout system
- Pro: Professional, reference-implementation approach
- Con: Requires understanding Ultraviolet's rectangle system

---

### Option B: Remove Ultraviolet (Your Original Plan)
```go
// Simple string composition - no canvas needed

func (m *UI) View() tea.View {
    v := tea.NewView("")
    v.AltScreen = false  // ❌ NOT fullscreen - loses polish
    
    // Just build one big string
    v.Content = strings.Join([]string{
        m.renderChat(),
        m.renderTools(),
        m.renderInput(),
    }, "\n")
    
    return v
}

// Result: More Unix-like but less polished
```

**Benefits:**
- Pro: Simpler codebase
- Pro: Works everywhere (even limited terminals)
- Con: Loses fullscreen UI elegance (Claude Code uses fullscreen!)
- Con: No per-pane layout control
- Con: Doesn't match Crush reference implementation
- Con: Less "polished" feeling that you want

---

### Option C: Hybrid (Best of Both Worlds)
```go
// Use Ultraviolet's layout WITHOUT the complex Draw() system

type Layout struct {
    ChatArea     uv.Rectangle
    StatusArea   uv.Rectangle
    InputArea    uv.Rectangle
}

func (m *UI) View() tea.View {
    v := tea.NewView("")
    v.AltScreen = true
    
    // Calculate layout once
    layout := m.calculateLayout()
    
    // Build strings for each section
    chatContent := m.renderChatInRegion(layout.ChatArea)
    statusContent := m.renderStatusInRegion(layout.StatusArea)
    inputContent := m.renderInputInRegion(layout.InputArea)
    
    // Compose with Lipgloss (no Ultraviolet rendering needed)
    v.Content = lipgloss.JoinVertical(
        lipgloss.Top,
        chatContent,
        statusContent,
        inputContent,
    )
    
    return v
}

// Result: Ultraviolet layout logic + simple string rendering
```

**Benefits:**
- Pro: High-level layout structure from Ultraviolet
- Pro: Simple string rendering (easy to understand)
- Pro: Keeps fullscreen UI + polished feel
- Pro: Per-region scrolling support
- Con: Mix of two approaches (but proven mix!)

---

## Recommendation Based on Crush Analysis

### Pick Option A or C (Keep Some Form of Layout Management)

**Why?**

1. **Crush validates it works** - Reference implementation uses Ultraviolet successfully
   
2. **Layout system matters** - Your requirement for "proper and perfect" design needs structured layout
   
3. **Per-pane scrolling is essential** 
   - Chat area scrolls ✓ 
   - Tool status updates dynamically ✓
   - Input prompt stays fixed ✓
   
4. **Police the implementation**, don't abandon the tool
   - Ghost's problem: 2200-line Draw() method ❌
   - Solution: Refactor into component-based Draw() like Crush ✓

---

## Crush's Actual Implementation Pattern (What You Should Copy)

```go
package ui

// 1. Define layout structure (not in Draw!)
type Layout struct {
    Main   uv.Rectangle
    Editor uv.Rectangle
    Help   uv.Rectangle
}

func GenerateLayout(width, height int) Layout {
    return Layout{
        Main:   uv.Rect(0, 0, width, height-4),
        Editor: uv.Rect(0, height-4, width, 3),
        Help:   uv.Rect(0, height-1, width, 1),
    }
}

// 2. Component-based Draw (NOT monolithic!)
type ChatComponent struct {
    messages []string
    viewport viewport.Model
}

func (c *ChatComponent) Draw(scr uv.Screen, area uv.Rectangle) {
    // Each component is responsible for its own region
    for i, line := range c.viewport.VisibleLines() {
        scr.DrawString(line, styleUser, area, i)
    }
}

// 3. Main UI delegates to components
type UI struct {
    chat   *ChatComponent
    editor *EditorComponent
    help   *HelpComponent
    layout Layout
}

func (m *UI) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
    screen.Clear(scr)
    m.chat.Draw(scr, m.layout.Main)      // ✓ Component owns region
    m.editor.Draw(scr, m.layout.Editor)  // ✓ Component owns region
    m.help.Draw(scr, m.layout.Help)      // ✓ Component owns region
    return nil
}

// 4. View just calls Draw (unchanged from current)
func (m *UI) View() tea.View {
    var v tea.View
    v.AltScreen = true
    canvas := uv.NewScreenBuffer(m.width, m.height)
    v.Cursor = m.Draw(canvas, canvas.Bounds())
    v.Content = canvas.Render()
    return v
}
```

**Result:** 
- Layout calculation: ~100 lines
- Each component's Draw: ~50-100 lines each
- Main UI.Draw: ~20 lines
- **Total ~500 lines instead of 2200**

---

## Decision Matrix

| Criterion | Option A (Fix UV) | Option B (Remove UV) | Option C (Hybrid) |
|-----------|------------------|---------------------|-------------------|
| Matches Crush pattern | ✅ Perfect | ❌ No | ✅ Very close |
| Fullscreen UI polish | ✅ Yes | ⚠️ No | ✅ Yes |
| Per-pane scrolling | ✅ Native | ❌ Manual | ✅ Easy |
| Code clarity | ✅ Good | ✅ Simpler | ✅ Good |
| Codebase size | ✅ ~500 lines | ✅ ~300 lines | ✅ ~400 lines |
| Reference impl | ✅ Crush uses it | ❌ Not standard | ✅ Crush pattern |
| Learning curve | ⚠️ Moderate | ✅ Easy | ✅ Easy |

---

## My Professional Recommendation

**Go with Option A: Fix the Ultraviolet Usage**

**Rationale:**

1. **Reference Implementation Exists** - Crush proves it can be done right

2. **Architecture Pattern is Clear**
   - Use layout rectangles for planning
   - Component-based Draw() methods
   - Each component owns its region
   
3. **Requirements Met**
   - ✅ Proper layout management
   - ✅ Fullscreen + polished
   - ✅ Per-component scrolling
   - ✅ Professional appearance
   - ✅ Proven pattern

4. **Refactoring Path is Clear**
   - Phase 1: Extract layout calculation from Draw()
   - Phase 2: Create ChatComponent with viewport
   - Phase 3: Create ToolComponent
   - Phase 4: Create InputComponent
   - Phase 5: Refactor main UI.Draw() to 20 lines
   
5. **Matches Your Vision**
   - You said "similar powerful like Claude Code"
   - Claude Code uses fullscreen TUI (like Crush)
   - Crush uses Ultraviolet (correctly)
   - Therefore: Ghost should too

---

## If You Still Want to Remove Ultraviolet

Use this reasoning:

**Ghost's specific use case:**
- Single chat stream (not multi-pane complex layout)
- Simple tool status display
- Natural terminal scrolling more valuable than fullscreen polish
- Team/users prefer Unix pipes and redirection behavior
- Scaling to 1000+ messages (string rendering sufficient)

Then: **Option B (Remove UV) is fine** - just own the decision to be less polished.

---

## Action Items

### If keeping Ultraviolet (Recommended):
1. Study Crush's `GenerateLayout()` implementation
2. Copy pattern of component-based Draw() methods
3. Extract layout early in Phase 1
4. Build viewport support for chat incrementally

### If removing Ultraviolet:
1. Accept trade-off: Less polished UI
2. Focus on string composition excellence
3. Update planning documents to reflect "Unix-first" philosophy
4. Test with pipes/tmux/ssh extensively

---

## Code Reference

**Crush's actual layout generation** (to copy/adapt):
- File: `internal/ui/model/ui.go` → `generateLayout()`
- Pattern: Calculate rectangles upfront, use in Draw()

**Crush's component Draw pattern** (to copy/adapt):
- File: `internal/ui/model/chat.go` → `Draw(scr, rect)`
- Pattern: Each component renders only in its rectangle

---

## TL;DR

| Point | Reality |
|-------|---------|
| Is Ultraviolet good? | ✅ Yes - Crush uses it |
| Should you remove it? | ❌ No - Fix your usage instead |
| What's the real problem? | Ghost doesn't use layout management properly |
| What should you do? | Copy Crush's component-based Draw() pattern |
| Result | Fullscreen, polished, 500-line UI model |

**The decision: REFACTOR, don't REMOVE.** 🎯
