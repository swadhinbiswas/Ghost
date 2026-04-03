package components

import (
tea "charm.land/bubbletea/v2"
uv "github.com/charmbracelet/ultraviolet"
"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

type HelpComponent struct {
	*BaseComponent
	styles *styles.Styles
}

func NewHelpComponent(s *styles.Styles) *HelpComponent {
	return &HelpComponent{
		BaseComponent: NewBaseComponent("help"),
		styles:        s,
	}
}

func (c *HelpComponent) Draw(scr uv.Screen, area uv.Rectangle) {
	if !c.visible {
		return
	}
	content := c.styles.Base.Render("Help Component (Ultraviolet rendered)")
	styled := uv.NewStyledString(content)
	styled.Draw(scr, area)
}

func (c *HelpComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}
