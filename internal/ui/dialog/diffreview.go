package dialog

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/sahilm/fuzzy"
	"github.com/swadhinbiswas/ghost/internal/ui/common"
	"github.com/swadhinbiswas/ghost/internal/ui/list"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

const (
	DiffReviewID              = "diff_review"
	diffReviewDialogMaxWidth  = 80
	diffReviewDialogMaxHeight = 25
)

// DiffHunk represents a single change block in a unified diff.
type DiffHunk struct {
	OldStart int
	NewStart int
	OldLines int
	NewLines int
	Content  string
	Accepted bool
	Rejected bool
}

// DiffFile represents all changes for a single file.
type DiffFile struct {
	FilePath string
	Hunks    []DiffHunk
}

// DiffReview represents a dialog for reviewing and selectively accepting diff changes.
type DiffReview struct {
	com       *common.Common
	help      help.Model
	list      *list.FilterableList
	input     textinput.Model
	files     []DiffFile
	hunkItems []list.FilterableItem

	keyMap struct {
		Toggle    key.Binding
		Next      key.Binding
		Previous  key.Binding
		AcceptAll key.Binding
		RejectAll key.Binding
		Confirm   key.Binding
		Close     key.Binding
	}
}

var _ Dialog = (*DiffReview)(nil)

// DiffReviewResult is returned when the user confirms or cancels the review.
type DiffReviewResult struct {
	AcceptedHunks map[string]string // filePath -> accepted diff content
	Cancelled     bool
}

// NewDiffReview creates a new diff review dialog.
func NewDiffReview(com *common.Common, files []DiffFile) *DiffReview {
	dr := &DiffReview{
		com:   com,
		files: files,
	}

	h := help.New()
	h.Styles = com.Styles.DialogHelpStyles()
	dr.help = h

	dr.list = list.NewFilterableList()
	dr.list.Focus()

	dr.input = textinput.New()
	dr.input.SetVirtualCursor(false)
	dr.input.Placeholder = "Filter hunks"
	dr.input.SetStyles(com.Styles.TextInput)
	dr.input.Focus()

	dr.hunkItems = dr.buildHunkItems()
	dr.list.SetItems(dr.hunkItems...)

	dr.keyMap.Toggle = key.NewBinding(
		key.WithKeys("space", "tab"),
		key.WithHelp("space", "toggle"),
	)
	dr.keyMap.Next = key.NewBinding(
		key.WithKeys("down", "ctrl+n"),
		key.WithHelp("↓", "next"),
	)
	dr.keyMap.Previous = key.NewBinding(
		key.WithKeys("up", "ctrl+p"),
		key.WithHelp("↑", "prev"),
	)
	dr.keyMap.AcceptAll = key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "accept all"),
	)
	dr.keyMap.RejectAll = key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reject all"),
	)
	dr.keyMap.Confirm = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "apply"),
	)
	dr.keyMap.Close = CloseKey

	return dr
}

// ID implements Dialog.
func (dr *DiffReview) ID() string {
	return DiffReviewID
}

// HandleMsg implements Dialog.
func (dr *DiffReview) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, dr.keyMap.Close):
			return DiffReviewResult{Cancelled: true}
		case key.Matches(msg, dr.keyMap.Toggle):
			dr.toggleCurrentHunk()
			return nil
		case key.Matches(msg, dr.keyMap.Next):
			dr.list.Focus()
			if dr.list.IsSelectedLast() {
				dr.list.SelectFirst()
				dr.list.ScrollToTop()
			} else {
				dr.list.SelectNext()
				dr.list.ScrollToSelected()
			}
			return nil
		case key.Matches(msg, dr.keyMap.Previous):
			dr.list.Focus()
			if dr.list.IsSelectedFirst() {
				dr.list.SelectLast()
				dr.list.ScrollToBottom()
			} else {
				dr.list.SelectPrev()
				dr.list.ScrollToSelected()
			}
			return nil
		case key.Matches(msg, dr.keyMap.AcceptAll):
			dr.acceptAll()
			return nil
		case key.Matches(msg, dr.keyMap.RejectAll):
			dr.rejectAll()
			return nil
		case key.Matches(msg, dr.keyMap.Confirm):
			return dr.buildResult()
		default:
			var cmd tea.Cmd
			dr.input, cmd = dr.input.Update(msg)
			dr.list.SetFilter(dr.input.Value())
			dr.list.ScrollToTop()
			dr.list.SetSelected(0)
			return ActionCmd{cmd}
		}
	}
	return nil
}

func (dr *DiffReview) toggleCurrentHunk() {
	idx := dr.list.Selected()
	filtered := dr.list.FilteredItems()
	if idx < 0 || idx >= len(filtered) {
		return
	}
	item, ok := filtered[idx].(*diffHunkItem)
	if !ok {
		return
	}
	// Find the actual hunk in dr.files and toggle it
	for fi := range dr.files {
		for hi := range dr.files[fi].Hunks {
			if &dr.files[fi].Hunks[hi] == item.hunk {
				h := &dr.files[fi].Hunks[hi]
				if h.Accepted {
					h.Accepted = false
					h.Rejected = true
				} else if h.Rejected {
					h.Accepted = false
					h.Rejected = false
				} else {
					h.Accepted = true
				}
				dr.list.SetItems(dr.buildHunkItems()...)
				return
			}
		}
	}
}

func (dr *DiffReview) acceptAll() {
	for fi := range dr.files {
		for hi := range dr.files[fi].Hunks {
			dr.files[fi].Hunks[hi].Accepted = true
			dr.files[fi].Hunks[hi].Rejected = false
		}
	}
	dr.list.SetItems(dr.buildHunkItems()...)
}

func (dr *DiffReview) rejectAll() {
	for fi := range dr.files {
		for hi := range dr.files[fi].Hunks {
			dr.files[fi].Hunks[hi].Accepted = false
			dr.files[fi].Hunks[hi].Rejected = true
		}
	}
	dr.list.SetItems(dr.buildHunkItems()...)
}

func (dr *DiffReview) buildResult() Action {
	result := DiffReviewResult{
		AcceptedHunks: make(map[string]string),
	}
	for _, f := range dr.files {
		var acceptedContent strings.Builder
		hasAccepted := false
		for _, h := range f.Hunks {
			if h.Accepted {
				acceptedContent.WriteString(h.Content)
				acceptedContent.WriteString("\n")
				hasAccepted = true
			}
		}
		if hasAccepted {
			result.AcceptedHunks[f.FilePath] = acceptedContent.String()
		}
	}
	return result
}

func (dr *DiffReview) buildHunkItems() []list.FilterableItem {
	items := []list.FilterableItem{}
	for _, f := range dr.files {
		for hi := range f.Hunks {
			h := &f.Hunks[hi]
			status := "pending"
			if h.Accepted {
				status = "accepted"
			} else if h.Rejected {
				status = "rejected"
			}
			items = append(items, &diffHunkItem{
				filePath: f.FilePath,
				hunkIdx:  hi,
				status:   status,
				hunk:     h,
				t:        dr.com.Styles,
			})
		}
	}
	return items
}

type diffHunkItem struct {
	filePath string
	hunkIdx  int
	status   string
	hunk     *DiffHunk
	t        *styles.Styles
	m        fuzzy.Match
	cache    map[int]string
	focused  bool
}

func (d *diffHunkItem) Filter() string {
	return d.filePath + " " + d.status
}

func (d *diffHunkItem) ID() string {
	return fmt.Sprintf("%s:%d", d.filePath, d.hunkIdx)
}

func (d *diffHunkItem) SetFocused(focused bool) {
	if d.focused != focused {
		d.cache = nil
	}
	d.focused = focused
}

func (d *diffHunkItem) SetMatch(m fuzzy.Match) {
	d.cache = nil
	d.m = m
}

func (d *diffHunkItem) Render(width int) string {
	label := fmt.Sprintf("%s (hunk %d)", d.filePath, d.hunkIdx+1)
	info := d.status

	st := ListItemStyles{
		ItemBlurred:     d.t.Dialog.NormalItem,
		ItemFocused:     d.t.Dialog.SelectedItem,
		InfoTextBlurred: d.t.Base,
		InfoTextFocused: d.t.Base,
	}

	if len(label) > width-10 {
		label = ansi.Truncate(label, width-10, "…")
	}

	return renderItem(st, label, info, d.focused, width, d.cache, &d.m)
}

// Draw implements Dialog.
func (dr *DiffReview) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := dr.com.Styles
	width := max(0, min(diffReviewDialogMaxWidth, area.Dx()))
	height := max(0, min(diffReviewDialogMaxHeight, area.Dy()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()
	heightOffset := t.Dialog.Title.GetVerticalFrameSize() + titleContentHeight +
		t.Dialog.InputPrompt.GetVerticalFrameSize() + inputContentHeight +
		t.Dialog.HelpView.GetVerticalFrameSize() +
		t.Dialog.View.GetVerticalFrameSize()

	dr.input.SetWidth(innerWidth - t.Dialog.InputPrompt.GetHorizontalFrameSize() - 1)
	dr.list.SetSize(innerWidth, height-heightOffset)
	dr.help.SetWidth(innerWidth)

	rc := NewRenderContext(t, width)
	rc.Title = "Review Changes"

	// Summary
	totalHunks := 0
	accepted := 0
	rejected := 0
	for _, f := range dr.files {
		for _, h := range f.Hunks {
			totalHunks++
			if h.Accepted {
				accepted++
			} else if h.Rejected {
				rejected++
			}
		}
	}
	summary := fmt.Sprintf("  %d hunks · %d accepted · %d rejected · %d pending",
		totalHunks, accepted, rejected, totalHunks-accepted-rejected)
	rc.AddPart(t.Base.Foreground(t.Secondary).Render(summary))

	// Diff preview for current hunk
	idx := dr.list.Selected()
	filtered := dr.list.FilteredItems()
	var diffView strings.Builder
	if idx >= 0 && idx < len(filtered) {
		if item, ok := filtered[idx].(*diffHunkItem); ok {
			diffView.WriteString(fmt.Sprintf("  File: %s (hunk %d)\n\n", item.filePath, item.hunkIdx+1))
			lines := strings.Split(item.hunk.Content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					diffView.WriteString(t.Base.Foreground(t.Green).Render("  "+line) + "\n")
				} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					diffView.WriteString(t.Base.Foreground(t.Red).Render("  "+line) + "\n")
				} else {
					diffView.WriteString(t.Base.Faint(true).Render("  "+line) + "\n")
				}
			}
		}
	}

	diffStr := ansi.Truncate(diffView.String(), width-4, "…")
	rc.AddPart(t.Base.Render(diffStr))

	// List
	visibleCount := len(dr.list.FilteredItems())
	if dr.list.Height() >= visibleCount {
		dr.list.ScrollToTop()
	} else {
		dr.list.ScrollToSelected()
	}
	listView := t.Dialog.List.Height(dr.list.Height()).Render(dr.list.Render())
	rc.AddPart(listView)

	rc.Help = dr.help.View(dr)
	view := rc.Render()

	cur := dr.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}

// Cursor returns the cursor position relative to the dialog.
func (dr *DiffReview) Cursor() *tea.Cursor {
	return InputCursor(dr.com.Styles, dr.input.Cursor())
}

// ShortHelp implements help.KeyMap.
func (dr *DiffReview) ShortHelp() []key.Binding {
	return []key.Binding{
		dr.keyMap.Toggle,
		dr.keyMap.Confirm,
		dr.keyMap.Close,
	}
}

// FullHelp implements help.KeyMap.
func (dr *DiffReview) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{dr.keyMap.Toggle, dr.keyMap.AcceptAll, dr.keyMap.RejectAll, dr.keyMap.Close},
		{dr.keyMap.Next, dr.keyMap.Previous, dr.keyMap.Confirm},
	}
}
