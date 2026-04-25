#!/bin/bash

# Remove uv from types.go
sed -i 's/uv "github.com\/charmbracelet\/ultraviolet"//g' internal/ui/components/types.go
sed -i 's/Draw(scr uv.Screen, area uv.Rectangle)/View() string/g' internal/ui/components/types.go
sed -i '/type DrawLayout struct {/,+17d' internal/ui/components/types.go

# Rewrite chat_component.go
cat << 'INNER_EOF' > internal/ui/components/chat_component.go
package components

import (
"charm.land/bubbletea/v2"
"github.com/charmbracelet/lipgloss"
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

func (c *ChatComponent) View() string {
	if !c.visible || c.height <= 0 || c.width <= 0 {
		return ""
	}
	content := lipgloss.NewStyle().Width(c.width).Height(c.height).Render("Chat Component (Lipgloss rendered)")
	return c.styles.Base.Render(content)
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
INNER_EOF

# Rewrite input_component.go
cat << 'INNER_EOF' > internal/ui/components/input_component.go
package components

import (
"image"
"charm.land/bubbletea/v2"
"github.com/charmbracelet/lipgloss"
"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

type InputComponent struct {
	*BaseComponent
	styles    *styles.Styles
	textarea  string
	cursorRow int
	cursorCol int
	focused   bool
}

var _ EditableComponent = (*InputComponent)(nil)
var _ FocusableComponent = (*InputComponent)(nil)

func NewInputComponent(s *styles.Styles) *InputComponent {
	return &InputComponent{
		BaseComponent: NewBaseComponent("input"),
		styles:        s,
		focused:       true,
	}
}

func (c *InputComponent) View() string {
	if !c.visible || c.height <= 0 || c.width <= 0 {
		return ""
	}
	prompt := c.styles.EditorPromptNormalFocused.Render("::: ")
	if !c.focused {
		prompt = c.styles.EditorPromptNormalBlurred.Render("::: ")
	}
	content := lipgloss.JoinHorizontal(lipgloss.Left, prompt, c.textarea)
	return lipgloss.NewStyle().Width(c.width).Height(c.height).Render(content)
}

func (c *InputComponent) Update(msg tea.Msg) (Component, tea.Cmd) { return c, nil }
func (c *InputComponent) Focus() { c.focused = true }
func (c *InputComponent) Blur() { c.focused = false }
func (c *InputComponent) IsFocused() bool { return c.focused }
func (c *InputComponent) SetText(text string) { c.textarea = text }
func (c *InputComponent) GetText() string { return c.textarea }
func (c *InputComponent) Clear() { c.textarea = ""; c.cursorCol = 0; c.cursorRow = 0 }
func (c *InputComponent) CursorPosition() image.Point { return image.Point{X: c.cursorCol, Y: c.cursorRow} }
INNER_EOF

# Rewrite tools_component.go
cat << 'INNER_EOF' > internal/ui/components/tools_component.go
package components

import (
"charm.land/bubbletea/v2"
"github.com/charmbracelet/lipgloss"
"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

type ToolsComponent struct {
	*BaseComponent
	styles      *styles.Styles
	activeTools []string
}

func NewToolsComponent(s *styles.Styles) *ToolsComponent {
	return &ToolsComponent{
		BaseComponent: NewBaseComponent("tools"),
		styles:        s,
		activeTools:   make([]string, 0),
	}
}

func (c *ToolsComponent) View() string {
	if !c.visible || len(c.activeTools) == 0 {
		return ""
	}
	
	var lines []string
	for _, t := range c.activeTools {
		lines = append(lines, c.styles.Base.Render(" ⏳ Running: "+t))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (c *ToolsComponent) Update(msg tea.Msg) (Component, tea.Cmd) { return c, nil }
INNER_EOF

# Fix Container Draw -> View in types.go
sed -i 's/func (c \*Container) Draw(scr uv.Screen, area uv.Rectangle) {/func (c \*Container) View() string {/g' internal/ui/components/types.go
sed -i '/if !c.visible {/,+8d' internal/ui/components/types.go
sed -i '/func (c \*Container) View() string {/a\
if !c.visible {\nreturn ""\n}\n\nvar views []string\nfor _, comp := range c.components {\nviews = append(views, comp.View())\n}\n// Default container stacks vertically - adjust as needed using lipgloss\nreturn lipgloss.JoinVertical(lipgloss.Left, views...)' internal/ui/components/types.go

# Fix BaseComponent Draw -> View in types.go
sed -i 's/func (c \*BaseComponent) Draw(scr uv.Screen, area uv.Rectangle) {/func (c \*BaseComponent) View() string {/g' internal/ui/components/types.go
sed -i 's/\/\/ Override in subclass/return ""/g' internal/ui/components/types.go

