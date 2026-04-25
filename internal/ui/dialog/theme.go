package dialog

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"
	"github.com/swadhinbiswas/ghost/internal/ui/common"
	"github.com/swadhinbiswas/ghost/internal/ui/list"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

const (
	// ThemePickerID is the identifier for the theme picker dialog.
	ThemePickerID        = "theme"
	themeDialogMaxWidth  = 62
	themeDialogMaxHeight = 18
	colorSwatchWidth     = 2
)

// ThemePicker represents a dialog for selecting a color theme.
type ThemePicker struct {
	com   *common.Common
	help  help.Model
	list  *list.FilterableList
	input textinput.Model

	keyMap struct {
		Select   key.Binding
		Next     key.Binding
		Previous key.Binding
		UpDown   key.Binding
		Close    key.Binding
	}
}

// ThemeItem represents a theme list item.
type ThemeItem struct {
	name      string
	isCurrent bool
	t         *styles.Styles
	theme     styles.Theme
	m         fuzzy.Match
	cache     map[int]string
	focused   bool
}

var (
	_ Dialog   = (*ThemePicker)(nil)
	_ ListItem = (*ThemeItem)(nil)
)

// NewThemePicker creates a new theme picker dialog.
func NewThemePicker(com *common.Common) *ThemePicker {
	tp := &ThemePicker{com: com}

	h := help.New()
	h.Styles = com.Styles.DialogHelpStyles()
	tp.help = h

	tp.list = list.NewFilterableList()
	tp.list.Focus()

	tp.input = textinput.New()
	tp.input.SetVirtualCursor(false)
	tp.input.Placeholder = "Find a theme"
	tp.input.SetStyles(com.Styles.TextInput)
	tp.input.Focus()

	tp.keyMap.Select = key.NewBinding(
		key.WithKeys("enter", "ctrl+y"),
		key.WithHelp("enter", "confirm"),
	)
	tp.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next"),
	)
	tp.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "previous"),
	)
	tp.keyMap.UpDown = key.NewBinding(
		key.WithKeys("up", "down"),
		key.WithHelp("↑/↓", "choose"),
	)
	tp.keyMap.Close = CloseKey

	tp.setThemeItems()

	return tp
}

// ID implements Dialog.
func (tp *ThemePicker) ID() string {
	return ThemePickerID
}

// HandleMsg implements [Dialog].
func (tp *ThemePicker) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, tp.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, tp.keyMap.Previous):
			tp.list.Focus()
			if tp.list.IsSelectedFirst() {
				tp.list.SelectLast()
				tp.list.ScrollToBottom()
				break
			}
			tp.list.SelectPrev()
			tp.list.ScrollToSelected()
		case key.Matches(msg, tp.keyMap.Next):
			tp.list.Focus()
			if tp.list.IsSelectedLast() {
				tp.list.SelectFirst()
				tp.list.ScrollToTop()
				break
			}
			tp.list.SelectNext()
			tp.list.ScrollToSelected()
		case key.Matches(msg, tp.keyMap.Select):
			selectedItem := tp.list.SelectedItem()
			if selectedItem == nil {
				break
			}
			themeItem, ok := selectedItem.(*ThemeItem)
			if !ok {
				break
			}
			return ActionSetTheme{Theme: themeItem.name}
		default:
			var cmd tea.Cmd
			tp.input, cmd = tp.input.Update(msg)
			value := tp.input.Value()
			tp.list.SetFilter(value)
			tp.list.ScrollToTop()
			tp.list.SetSelected(0)
			return ActionCmd{cmd}
		}
	}
	return nil
}

// Cursor returns the cursor position relative to the dialog.
func (tp *ThemePicker) Cursor() *tea.Cursor {
	return InputCursor(tp.com.Styles, tp.input.Cursor())
}

// Draw implements [Dialog].
func (tp *ThemePicker) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := tp.com.Styles
	width := max(0, min(themeDialogMaxWidth, area.Dx()))
	height := max(0, min(themeDialogMaxHeight, area.Dy()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()

	tp.input.SetWidth(innerWidth - t.Dialog.InputPrompt.GetHorizontalFrameSize() - 1)
	tp.list.SetSize(innerWidth, height-heightOffset)
	tp.help.SetWidth(innerWidth)

	rc := NewRenderContext(t, width)
	rc.Title = "Switch Theme"
	inputView := t.Dialog.InputPrompt.Render(tp.input.View())
	rc.AddPart(inputView)

	visibleCount := len(tp.list.FilteredItems())
	if tp.list.Height() >= visibleCount {
		tp.list.ScrollToTop()
	} else {
		tp.list.ScrollToSelected()
	}

	listView := t.Dialog.List.Height(tp.list.Height()).Render(tp.list.Render())
	rc.AddPart(listView)
	rc.Help = tp.help.View(tp)

	view := rc.Render()

	cur := tp.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// ShortHelp implements [help.KeyMap].
func (tp *ThemePicker) ShortHelp() []key.Binding {
	return []key.Binding{
		tp.keyMap.UpDown,
		tp.keyMap.Select,
		tp.keyMap.Close,
	}
}

// FullHelp implements [help.KeyMap].
func (tp *ThemePicker) FullHelp() [][]key.Binding {
	m := [][]key.Binding{}
	slice := []key.Binding{
		tp.keyMap.Select,
		tp.keyMap.Next,
		tp.keyMap.Previous,
		tp.keyMap.Close,
	}
	for i := 0; i < len(slice); i += 4 {
		end := min(i+4, len(slice))
		m = append(m, slice[i:end])
	}
	return m
}

func (tp *ThemePicker) setThemeItems() {
	themeNames := styles.ThemeNames()
	currentTheme := tp.com.App.GetTheme()

	items := make([]list.FilterableItem, 0, len(themeNames))
	selectedIndex := 0
	for i, name := range themeNames {
		item := &ThemeItem{
			name:      name,
			isCurrent: name == currentTheme,
			t:         tp.com.Styles,
			theme:     styles.GetTheme(name),
		}
		items = append(items, item)
		if name == currentTheme {
			selectedIndex = i
		}
	}

	tp.list.SetItems(items...)
	tp.list.SetSelected(selectedIndex)
	tp.list.ScrollToSelected()
}

// Filter returns the filter value for the theme item.
func (ti *ThemeItem) Filter() string {
	return ti.name
}

// ID returns the unique identifier for the theme.
func (ti *ThemeItem) ID() string {
	return ti.name
}

// SetFocused sets the focus state of the theme item.
func (ti *ThemeItem) SetFocused(focused bool) {
	if ti.focused != focused {
		ti.cache = nil
	}
	ti.focused = focused
}

// SetMatch sets the fuzzy match for the theme item.
func (ti *ThemeItem) SetMatch(m fuzzy.Match) {
	ti.cache = nil
	ti.m = m
}

// colorSwatch renders a small block showing the theme's key colors.
func (ti *ThemeItem) colorSwatch() string {
	bgBlock := lipgloss.NewStyle().Background(ti.theme.BgBase).Foreground(ti.theme.BgBase).Render("  ")
	primaryBlock := lipgloss.NewStyle().Background(ti.theme.Primary).Foreground(ti.theme.Primary).Render("  ")
	secondaryBlock := lipgloss.NewStyle().Background(ti.theme.Secondary).Foreground(ti.theme.Secondary).Render("  ")
	return bgBlock + primaryBlock + secondaryBlock
}

// Render returns the string representation of the theme item.
func (ti *ThemeItem) Render(width int) string {
	swatch := ti.colorSwatch()
	swatchWidth := lipgloss.Width(swatch)
	labelWidth := width - swatchWidth - 2

	info := ""
	if ti.isCurrent {
		info = "active"
	}
	st := ListItemStyles{
		ItemBlurred:     ti.t.Dialog.NormalItem,
		ItemFocused:     ti.t.Dialog.SelectedItem,
		InfoTextBlurred: ti.t.Base,
		InfoTextFocused: ti.t.Base,
	}

	nameView := ti.name
	if lipgloss.Width(nameView) > labelWidth {
		nameView = ansi.Truncate(nameView, labelWidth, "…")
	}

	label := lipgloss.NewStyle().Width(labelWidth).Render(nameView)
	content := label + " " + swatch

	itemStyle := st.ItemBlurred
	if ti.focused {
		itemStyle = st.ItemFocused
	}

	line := itemStyle.Render(content)

	if info != "" {
		infoStyle := st.InfoTextBlurred
		if ti.focused {
			infoStyle = st.InfoTextFocused
		}
		line = lipgloss.JoinHorizontal(lipgloss.Center, line, infoStyle.Render(info))
	}

	return line
}
