package components

import (
	"image"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

type InputComponent struct {
	*BaseComponent
	textarea textarea.Model
	styles   *styles.Styles
}

var _ EditableComponent = (*InputComponent)(nil)

func NewInputComponent(s *styles.Styles) *InputComponent {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.Focus()

	return &InputComponent{
		BaseComponent: NewBaseComponent("input"),
		textarea:      ta,
		styles:        s,
	}
}

func (c *InputComponent) Draw(scr uv.Screen, area uv.Rectangle) {
	if !c.visible {
		return
	}

	c.textarea.SetWidth(area.Dx())
	c.textarea.SetHeight(area.Dy())

	styled := uv.NewStyledString(c.textarea.View())
	styled.Draw(scr, area)
}

func (c *InputComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)
	return c, cmd
}

func (c *InputComponent) SetText(text string) {
	c.textarea.SetValue(text)
}

func (c *InputComponent) GetText() string {
	return strings.TrimSpace(c.textarea.Value())
}

func (c *InputComponent) Clear() {
	c.textarea.Reset()
}

func (c *InputComponent) CursorPosition() image.Point {
	return image.Pt(0, 0)
}
