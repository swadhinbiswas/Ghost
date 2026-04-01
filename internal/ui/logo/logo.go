package logo

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

const diag = "╱"

// Dracula Theme Colors
var (
	colorPurple = lipgloss.Color("#bb9af7")
	colorPink   = lipgloss.Color("#ff4d88")
	colorDark   = lipgloss.Color("#0b0b10")
	colorFG     = lipgloss.Color("#f8f8f2")
)

type Opts struct {
	FieldColor   color.Color
	TitleColorA  color.Color
	TitleColorB  color.Color
	CharmColor   color.Color
	VersionColor color.Color
	Width        int
}

var ghostLines = []string{
	"⠀⠀⠄⠀⠀⠂⠀⠀⠀⡀⠀⠀",
	"⠁⠀⠀⣠⣶⣿⣷⣶⣄⠀⠀⠁",
	"⠈⠀⢰⡿⠛⢿⡿⠻⣿⡆⠀⡀",
	"⠀⠠⣾⡇⠀⢸⡇⠀⢸⣧⠀⠀",
	"⠀⠀⣿⣷⣤⣿⣷⣤⣿⣿⠀⠀",
	"⠠⠀⣿⣿⣿⣿⣿⣿⣿⣿⠀⡀",
	"⠄⡀⠙⠟⢿⣿⡿⠿⠿⠋⠀⠀",
}

func Render(s *styles.Styles, version string, compact bool, o Opts) string {
	purple := lipgloss.NewStyle().Foreground(colorPurple)
	pink := lipgloss.NewStyle().Foreground(colorPink)
	fg := lipgloss.NewStyle().Foreground(colorFG)

	rendered := make([]string, len(ghostLines))
	for i, line := range ghostLines {
		switch i {
		case 0:
			rendered[i] = pink.Render(line)
		case 1, 2:
			rendered[i] = purple.Render(line)
		case 3, 4:
			rendered[i] = purple.Render(line)
		case 5:
			rendered[i] = purple.Bold(true).Render(line)
		case 6:
			rendered[i] = fg.Render(line)
		}
	}

	logo := strings.Join(rendered, "\\n")

	if version != "" {
		ver := lipgloss.NewStyle().
			Foreground(colorPink).
			Faint(true).
			Render("v" + version)
		logo = fmt.Sprintf("%s  %s", logo, ver)
	}

	return logo
}

func SmallRender(t *styles.Styles, width int) string {
	title := t.Base.Foreground(t.Secondary).Render("Ghost")
	title = fmt.Sprintf("%s %s", title, styles.ApplyBoldForegroundGrad(t, "Ghost", t.Secondary, t.Primary))
	remainingWidth := width - lipgloss.Width(title) - 1
	if remainingWidth > 0 {
		lines := strings.Repeat(diag, remainingWidth)
		title = fmt.Sprintf("%s %s", title, t.Base.Foreground(t.Primary).Render(lines))
	}

	return title
}
func RenderAnimated(s *styles.Styles, version string, compact bool, o Opts, frame int) string {
	purple := lipgloss.NewStyle().Foreground(colorPurple)
	pink := lipgloss.NewStyle().Foreground(colorPink)
	fg := lipgloss.NewStyle().Foreground(colorFG)

	// Give feeling of floating up/down
	floatOffset := frame % 4
	if floatOffset == 3 {
		floatOffset = 1
	}

	rendered := make([]string, len(ghostLines)+2)
	topPadding := floatOffset
	bottomPadding := 2 - topPadding

	for i := 0; i < topPadding; i++ {
		rendered[i] = "                     "
	}

	idx := topPadding
	for i, line := range ghostLines {
		// blink eyes
		if i == 3 && frame%10 < 2 {
			line = strings.Replace(line, "⣿", "⠛", 2) // just fake blinking
		}

		switch i {
		case 0:
			// particles changing colors
			rendered[idx] = pink.Foreground(colorPurple).Render(line)
		case 1, 2:
			rendered[idx] = purple.Render(line)
		case 3, 4:
			rendered[idx] = purple.Render(line)
		case 5:
			rendered[idx] = purple.Bold(true).Render(line)
		case 6:
			rendered[idx] = fg.Render(line)
		}
		idx++
	}

	for i := 0; i < bottomPadding; i++ {
		rendered[idx] = "                     "
		idx++
	}

	logo := strings.Join(rendered, "\n")

	if version != "" {
		ver := lipgloss.NewStyle().
			Foreground(colorPink).
			Faint(true).
			Render("v" + version)
		logo = fmt.Sprintf("%s  %s", logo, ver)
	}

	return logo
}
