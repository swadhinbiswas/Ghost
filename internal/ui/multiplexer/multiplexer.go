package multiplexer

import (
	"fmt"
	"image"
	"strings"

	"charm.land/bubbles/v2/key"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

// PaneLayout defines how panes are arranged.
type PaneLayout string

const (
	LayoutHorizontal PaneLayout = "horizontal"
	LayoutVertical   PaneLayout = "vertical"
	LayoutGrid       PaneLayout = "grid"
	LayoutSingle     PaneLayout = "single"
)

// Pane represents a single panel in the multiplexer.
type Pane struct {
	ID      string
	Title   string
	Content string
	Focused bool
	Visible bool
	CursorX int
	CursorY int
	ScrollX int
	ScrollY int
}

// Multiplexer manages multiple terminal panes with layouts.
type Multiplexer struct {
	panes     []*Pane
	activeIdx int
	layout    PaneLayout
	width     int
	height    int
	styles    *styles.Styles

	keyMap struct {
		NextPane     key.Binding
		PrevPane     key.Binding
		SplitH       key.Binding
		SplitV       key.Binding
		ClosePane    key.Binding
		ToggleLayout key.Binding
		FocusMain    key.Binding
	}
}

// New creates a new multiplexer.
func New(styles *styles.Styles) *Multiplexer {
	m := &Multiplexer{
		layout:    LayoutSingle,
		styles:    styles,
		activeIdx: 0,
	}

	m.keyMap.NextPane = key.NewBinding(
		key.WithKeys("ctrl+]"),
		key.WithHelp("ctrl+]", "next pane"),
	)
	m.keyMap.PrevPane = key.NewBinding(
		key.WithKeys("ctrl+["),
		key.WithHelp("ctrl+[", "prev pane"),
	)
	m.keyMap.SplitH = key.NewBinding(
		key.WithKeys("ctrl+b", "h"),
		key.WithHelp("ctrl+b", "split horizontal"),
	)
	m.keyMap.SplitV = key.NewBinding(
		key.WithKeys("ctrl+v", "v"),
		key.WithHelp("ctrl+v", "split vertical"),
	)
	m.keyMap.ClosePane = key.NewBinding(
		key.WithKeys("ctrl+w", "x"),
		key.WithHelp("ctrl+w", "close pane"),
	)
	m.keyMap.ToggleLayout = key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "toggle layout"),
	)
	m.keyMap.FocusMain = key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "focus main"),
	)

	return m
}

// SetSize sets the overall size of the multiplexer.
func (m *Multiplexer) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// CreatePane creates a new pane and returns its ID.
func (m *Multiplexer) CreatePane(title, content string) string {
	id := fmt.Sprintf("pane-%d", len(m.panes)+1)
	pane := &Pane{
		ID:      id,
		Title:   title,
		Content: content,
		Visible: true,
	}
	m.panes = append(m.panes, pane)
	m.activeIdx = len(m.panes) - 1

	if len(m.panes) == 1 {
		m.layout = LayoutSingle
	} else if m.layout == LayoutSingle {
		m.layout = LayoutHorizontal
	}

	return id
}

// GetPane returns the pane with the given ID.
func (m *Multiplexer) GetPane(id string) *Pane {
	for _, p := range m.panes {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// ActivePane returns the currently focused pane.
func (m *Multiplexer) ActivePane() *Pane {
	if len(m.panes) == 0 {
		return nil
	}
	if m.activeIdx >= len(m.panes) {
		m.activeIdx = len(m.panes) - 1
	}
	return m.panes[m.activeIdx]
}

// SetActive sets the active pane by index.
func (m *Multiplexer) SetActive(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.panes) {
		idx = len(m.panes) - 1
	}
	m.activeIdx = idx
	m.updateFocus()
}

// NextPane moves focus to the next pane.
func (m *Multiplexer) NextPane() {
	if len(m.panes) <= 1 {
		return
	}
	m.activeIdx = (m.activeIdx + 1) % len(m.panes)
	m.updateFocus()
}

// PrevPane moves focus to the previous pane.
func (m *Multiplexer) PrevPane() {
	if len(m.panes) <= 1 {
		return
	}
	m.activeIdx = (m.activeIdx - 1 + len(m.panes)) % len(m.panes)
	m.updateFocus()
}

// ClosePane closes the active pane.
func (m *Multiplexer) ClosePane() {
	if len(m.panes) == 0 {
		return
	}

	// Don't close the last pane
	if len(m.panes) == 1 {
		return
	}

	m.panes = append(m.panes[:m.activeIdx], m.panes[m.activeIdx+1:]...)
	if m.activeIdx >= len(m.panes) {
		m.activeIdx = len(m.panes) - 1
	}
	m.updateFocus()

	if len(m.panes) == 1 {
		m.layout = LayoutSingle
	}
}

// ToggleLayout cycles through layouts.
func (m *Multiplexer) ToggleLayout() {
	layouts := []PaneLayout{LayoutSingle, LayoutHorizontal, LayoutVertical, LayoutGrid}
	for i, l := range layouts {
		if l == m.layout {
			m.layout = layouts[(i+1)%len(layouts)]
			break
		}
	}
}

// SetLayout sets the layout directly.
func (m *Multiplexer) SetLayout(layout PaneLayout) {
	m.layout = layout
}

// Layout returns the current layout.
func (m *Multiplexer) Layout() PaneLayout {
	return m.layout
}

// PaneCount returns the number of panes.
func (m *Multiplexer) PaneCount() int {
	return len(m.panes)
}

// CalculateAreas computes the rectangle for each pane based on the layout.
func (m *Multiplexer) CalculateAreas() []uv.Rectangle {
	if len(m.panes) == 0 {
		return nil
	}

	totalArea := uv.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: m.width, Y: m.height},
	}

	if m.layout == LayoutSingle || len(m.panes) == 1 {
		return []uv.Rectangle{totalArea}
	}

	areas := make([]uv.Rectangle, len(m.panes))
	n := len(m.panes)

	switch m.layout {
	case LayoutHorizontal:
		// Split horizontally (side by side)
		paneWidth := m.width / n
		for i := range n {
			x := i * paneWidth
			w := paneWidth
			if i == n-1 {
				w = m.width - x // Last pane takes remaining space
			}
			areas[i] = uv.Rectangle{
				Min: image.Point{X: x, Y: 0},
				Max: image.Point{X: x + w, Y: m.height},
			}
		}

	case LayoutVertical:
		// Split vertically (stacked)
		paneHeight := m.height / n
		for i := range n {
			y := i * paneHeight
			h := paneHeight
			if i == n-1 {
				h = m.height - y // Last pane takes remaining space
			}
			areas[i] = uv.Rectangle{
				Min: image.Point{X: 0, Y: y},
				Max: image.Point{X: m.width, Y: y + h},
			}
		}

	case LayoutGrid:
		// Grid layout (2x2 for 4 panes, etc.)
		cols := 2
		if n > 4 {
			cols = 3
		} else if n <= 2 {
			cols = n
		}
		rows := (n + cols - 1) / cols

		paneWidth := m.width / cols
		paneHeight := m.height / rows

		for i := range n {
			col := i % cols
			row := i / cols

			x := col * paneWidth
			y := row * paneHeight
			w := paneWidth
			h := paneHeight

			// Adjust last column/row to fill remaining space
			if col == cols-1 {
				w = m.width - x
			}
			if row == rows-1 {
				h = m.height - y
			}

			areas[i] = uv.Rectangle{
				Min: image.Point{X: x, Y: y},
				Max: image.Point{X: x + w, Y: y + h},
			}
		}
	}

	return areas
}

// Draw renders the multiplexer to the screen.
func (m *Multiplexer) Draw(scr uv.Screen, area uv.Rectangle) {
	if len(m.panes) == 0 {
		return
	}

	m.width = area.Dx()
	m.height = area.Dy()

	areas := m.CalculateAreas()
	t := m.styles

	for i, pane := range m.panes {
		if !pane.Visible || i >= len(areas) {
			continue
		}

		rect := areas[i]
		// Offset relative to the multiplexer area
		rect.Min.X += area.Min.X
		rect.Max.X += area.Min.X
		rect.Min.Y += area.Min.Y
		rect.Max.Y += area.Min.Y

		content := m.renderPane(pane, rect.Dx(), rect.Dy(), t)
		uv.NewStyledString(content).Draw(scr, rect)
	}
}

// renderPane renders a single pane with title bar and content.
func (m *Multiplexer) renderPane(pane *Pane, width, height int, t *styles.Styles) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	var sb strings.Builder

	// Title bar
	titleStyle := t.Base
	if pane.Focused {
		titleStyle = t.Base.Foreground(t.Primary).Bold(true)
	}

	titleBar := fmt.Sprintf(" %s ", pane.Title)
	titleBar = padOrTruncate(titleBar, width)
	sb.WriteString(titleStyle.Render(titleBar) + "\n")

	// Content
	contentHeight := height - 1
	if contentHeight <= 0 {
		return sb.String()
	}

	lines := strings.Split(pane.Content, "\n")

	// Apply scroll offset
	startLine := pane.ScrollY
	if startLine >= len(lines) {
		startLine = 0
	}
	lines = lines[startLine:]

	// Render visible lines
	for i := 0; i < contentHeight && i < len(lines); i++ {
		line := lines[i]
		line = padOrTruncate(line, width)
		sb.WriteString(line + "\n")
	}

	// Fill remaining space
	remaining := contentHeight - len(lines)
	if remaining > 0 {
		emptyLine := strings.Repeat(" ", width)
		for range remaining {
			sb.WriteString(emptyLine + "\n")
		}
	}

	return sb.String()
}

// ScrollPane scrolls the active pane.
func (m *Multiplexer) ScrollPane(dx, dy int) {
	pane := m.ActivePane()
	if pane == nil {
		return
	}
	pane.ScrollX += dx
	if pane.ScrollX < 0 {
		pane.ScrollX = 0
	}
	pane.ScrollY += dy
	if pane.ScrollY < 0 {
		pane.ScrollY = 0
	}
}

// UpdateContent updates the content of a pane.
func (m *Multiplexer) UpdateContent(id, content string) {
	pane := m.GetPane(id)
	if pane != nil {
		pane.Content = content
	}
}

// AppendContent appends content to a pane.
func (m *Multiplexer) AppendContent(id, content string) {
	pane := m.GetPane(id)
	if pane != nil {
		if pane.Content != "" {
			pane.Content += "\n" + content
		} else {
			pane.Content = content
		}
		// Auto-scroll to bottom
		pane.ScrollY = strings.Count(pane.Content, "\n")
	}
}

// KeyBindings returns the key bindings for help display.
func (m *Multiplexer) KeyBindings() []key.Binding {
	if len(m.panes) <= 1 {
		return nil
	}
	return []key.Binding{
		m.keyMap.NextPane,
		m.keyMap.PrevPane,
		m.keyMap.ClosePane,
		m.keyMap.ToggleLayout,
	}
}

func (m *Multiplexer) updateFocus() {
	for i, p := range m.panes {
		p.Focused = i == m.activeIdx
	}
}

// padOrTruncate ensures a string is exactly width characters.
func padOrTruncate(s string, width int) string {
	w := 0
	for range s {
		w++
	}
	if w > width {
		// Simple truncation (doesn't handle wide chars perfectly but works for most cases)
		if len(s) > width {
			return s[:width-1] + "…"
		}
	}
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}
