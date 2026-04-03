package components

import (
tea "charm.land/bubbletea/v2"
uv "github.com/charmbracelet/ultraviolet"
"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

type ToolsComponent struct {
	*BaseComponent
	styles *styles.Styles
}

func NewToolsComponent(s *styles.Styles) *ToolsComponent {
	return &ToolsComponent{
		BaseComponent: NewBaseComponent("tools"),
		styles:        s,
	}
}

func (c *ToolsComponent) Draw(scr uv.Screen, area uv.Rectangle) {
	if !c.visible {
		return
	}
	content := c.styles.Base.Render("Tools Component (Ultraviolet rendered)")
	styled := uv.NewStyledString(content)
	styled.Draw(scr, area)
}

func (c *ToolsComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}
