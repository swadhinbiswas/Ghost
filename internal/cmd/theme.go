package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
	"github.com/swadhinbiswas/ghost/internal/config"
	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Select a color theme",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		cfg, err := config.Load(cwd, "", false)
		if err != nil {
			return err
		}

		currentTheme := "Default"
		if cfg.Config().Options != nil && cfg.Config().Options.TUI != nil && cfg.Config().Options.TUI.Theme != "" {
			currentTheme = cfg.Config().Options.TUI.Theme
		}

		m := &themeModel{
			themes:       styles.ThemeNames(),
			currentTheme: currentTheme,
			config:       cfg,
		}

		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(themeCmd)
}

type themeModel struct {
	themes       []string
	cursor       int
	currentTheme string
	config       *config.ConfigStore
	saved        bool
}

func (m *themeModel) Init() tea.Cmd {
	return nil
}

func (m *themeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.themes) - 1
			}
		case "down", "j":
			if m.cursor < len(m.themes)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter", " ":
			// Save the theme
			m.currentTheme = m.themes[m.cursor]
			_ = m.config.SetTheme(config.ScopeGlobal, m.currentTheme)
			m.saved = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *themeModel) View() tea.View {
	if m.saved {
		t := styles.GetTheme(m.currentTheme)
		check := lipgloss.NewStyle().Foreground(t.Green).Bold(true).Render("✓")
		name := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(m.currentTheme)
		return tea.NewView(fmt.Sprintf("\n  %s Theme saved: %s\n\n", check, name))
	}

	s := strings.Builder{}
	s.WriteString("\n")

	// Title with gradient feel
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#9d80ff"))
	s.WriteString(fmt.Sprintf("  %s\n", titleStyle.Render("👻 Ghost Theme Selector")))

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#808080"))
	s.WriteString(fmt.Sprintf("  %s\n\n", helpStyle.Render("↑/↓ navigate • enter save • q quit")))

	for i, tName := range m.themes {
		t := styles.GetTheme(tName)

		// Color badges — show the full palette
		badges := []string{
			lipgloss.NewStyle().Background(t.Primary).Foreground(t.BgBase).Render(" Pr "),
			lipgloss.NewStyle().Background(t.Secondary).Foreground(t.BgBase).Render(" Sc "),
			lipgloss.NewStyle().Background(t.Tertiary).Foreground(t.BgBase).Render(" Tr "),
			lipgloss.NewStyle().Background(t.Error).Foreground(t.BgBase).Render(" Er "),
			lipgloss.NewStyle().Background(t.Warning).Foreground(t.BgBase).Render(" Wn "),
			lipgloss.NewStyle().Background(t.Info).Foreground(t.BgBase).Render(" In "),
			lipgloss.NewStyle().Background(t.Green).Foreground(t.BgBase).Render(" Gn "),
		}
		badgeStr := strings.Join(badges, "")

		// Cursor and active indicator
		cursor := "  "
		if m.cursor == i {
			cursor = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("> ")
		}

		// Theme name styling
		var label string
		if m.cursor == i {
			label = lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render(fmt.Sprintf("%-20s", tName))
		} else if tName == m.currentTheme {
			label = lipgloss.NewStyle().Foreground(t.Secondary).Render(fmt.Sprintf("%-20s", tName))
		} else {
			label = lipgloss.NewStyle().Foreground(lipgloss.Color("#999999")).Render(fmt.Sprintf("%-20s", tName))
		}

		// Active and light markers
		markers := ""
		if tName == m.currentTheme {
			markers += lipgloss.NewStyle().Foreground(t.Green).Bold(true).Render(" ✓")
		}
		if t.IsLight {
			markers += lipgloss.NewStyle().Foreground(t.Warning).Render(" ☀")
		}

		// Mini sample text showing what styled content looks like
		sample := ""
		if m.cursor == i {
			kw := lipgloss.NewStyle().Foreground(t.Blue).Render("func")
			fn := lipgloss.NewStyle().Foreground(t.GreenDark).Render("main")
			str := lipgloss.NewStyle().Foreground(t.Secondary).Render("\"hello\"")
			cmt := lipgloss.NewStyle().Foreground(t.FgSubtle).Render("// ok")
			sample = fmt.Sprintf("  %s %s(%s) %s", kw, fn, str, cmt)
		}

		s.WriteString(fmt.Sprintf("%s%s %s%s%s\n", cursor, label, badgeStr, markers, sample))
	}

	s.WriteString("\n")
	return tea.NewView(s.String())
}
