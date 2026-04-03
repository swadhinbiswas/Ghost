package components

import (
tea "charm.land/bubbletea/v2"
uv "github.com/charmbracelet/ultraviolet"
"github.com/swadhinbiswas/ghost/internal/ui/chat"
"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

type ChatComponent struct {
	*BaseComponent
	messages []chat.MessageItem
	styles   *styles.Styles
	scroll   int
}

var _ ScrollableComponent = (*ChatComponent)(nil)

func NewChatComponent(s *styles.Styles) *ChatComponent {
	return &ChatComponent{
		BaseComponent: NewBaseComponent("chat"),
		styles:        s,
		messages:      make([]chat.MessageItem, 0),
		scroll:        0,
	}
}

func (c *ChatComponent) SetMessages(messages []chat.MessageItem) {
	c.messages = messages
}

func (c *ChatComponent) Draw(scr uv.Screen, area uv.Rectangle) {
	if !c.visible {
		return
	}
	content := c.styles.Base.Render("Chat Component (Ultraviolet rendered)")
	styled := uv.NewStyledString(content)
	styled.Draw(scr, area)
}

func (c *ChatComponent) Update(msg tea.Msg) (Component, tea.Cmd) {
	return c, nil
}

func (c *ChatComponent) ScrollUp(n int) { c.scroll -= n; if c.scroll < 0 { c.scroll = 0 } }
func (c *ChatComponent) ScrollDown(n int) { c.scroll += n }
func (c *ChatComponent) ScrollToTop() { c.scroll = 0 }
func (c *ChatComponent) ScrollToBottom() {}
func (c *ChatComponent) IsAtTop() bool { return c.scroll == 0 }
func (c *ChatComponent) IsAtBottom() bool { return false }
