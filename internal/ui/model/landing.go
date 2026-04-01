package model

import (

	"time"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/swadhinbiswas/ghost/internal/agent"
	"github.com/swadhinbiswas/ghost/internal/ui/common"
	"github.com/swadhinbiswas/ghost/internal/ui/logo"
)

// selectedLargeModel returns the currently selected large language model from
// the agent coordinator, if one exists.
func (m *UI) selectedLargeModel() *agent.Model {
	if m.com.App.AgentCoordinator != nil {
		model := m.com.App.AgentCoordinator.Model()
		return &model
	}
	return nil
}

// landingView renders the landing page view showing the current working
// directory and keyboard shortcuts.
func (m *UI) landingView() string {
        t := m.com.Styles
        width := m.layout.main.Dx()
        height := m.layout.main.Dy() - 1
        
        cwd := common.PrettyPath(t, m.com.Store().WorkingDir(), width)

        title := logo.RenderAnimated(t, "", false, logo.Opts{}, m.landingFrame)
        
        infoSection := lipgloss.JoinVertical(lipgloss.Center, 
                title,
                "",
                cwd,
        )
        
        infoSection = lipgloss.PlaceHorizontal(width, lipgloss.Center, infoSection)

        shortcutWidth := min(60, width-10)
        shortcutWidth = max(40, shortcutWidth)

        // Create a beautifully formatted shortcuts view
        keyStyle := t.Base.Foreground(t.Primary).Bold(true)
        descStyle := t.Base.Foreground(t.FgSubtle)

        shortcutsList := []string{
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+p"), descStyle.MarginLeft(2).Render("commands")),
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+l"), descStyle.MarginLeft(2).Render("models")),
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+s"), descStyle.MarginLeft(2).Render("sessions")),
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+n"), descStyle.MarginLeft(2).Render("new session")),
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+f"), descStyle.MarginLeft(2).Render("add file/image")),
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+o"), descStyle.MarginLeft(2).Render("open editor")),
                lipgloss.JoinHorizontal(lipgloss.Left, keyStyle.Render("ctrl+x"), descStyle.MarginLeft(2).Render("exit")),
        }

        half := (len(shortcutsList) + 1) / 2
        col1 := lipgloss.JoinVertical(lipgloss.Left, shortcutsList[:half]...)
        col2 := lipgloss.JoinVertical(lipgloss.Left, shortcutsList[half:]...)

        shortcutsGrid := lipgloss.JoinHorizontal(lipgloss.Top, 
                col1, 
                lipgloss.NewStyle().Width(8).Render(""), 
                col2,
        )

        shortcutBox := lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(t.Border).
                Padding(1, 3).
                Width(shortcutWidth).
                Render(lipgloss.JoinVertical(lipgloss.Center, t.Base.Bold(true).Foreground(t.Secondary).Render("Keyboard Shortcuts"), "", shortcutsGrid))

        content := lipgloss.PlaceHorizontal(width, lipgloss.Center, shortcutBox)

        finalView := lipgloss.JoinVertical(lipgloss.Center, infoSection, "", "", "", content)
        
        return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, finalView)
}

func (m *UI) tickLanding() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {
		return tickLandingMsg(t)
	})
}
