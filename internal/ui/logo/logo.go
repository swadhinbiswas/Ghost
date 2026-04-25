package logo

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/swadhinbiswas/ghost/internal/ui/styles"
)

const diag = "╱"

type Opts struct {
	FieldColor   color.Color
	TitleColorA  color.Color
	TitleColorB  color.Color
	CharmColor   color.Color
	VersionColor color.Color
	Width        int
}

var ghostLines = []string{
	` ________  ___  ___  ________  ________  _________   `,
	`|\   ____\|\  \|\  \|\   __  \|\   ____\|\___   ___\ `,
	`\ \  \___|\ \  \\\  \ \  \|\  \ \  \___|\|___ \  \_| `,
	` \ \  \  __\ \   __  \ \  \\\  \ \_____  \   \ \  \  `,
	`  \ \  \|\  \ \  \ \  \ \  \\\  \|____|\  \   \ \  \ `,
	`   \ \_______\ \__\ \__\ \_______\____\_\  \   \ \__\`,
	`    \|_______|\|__|\|__|\|_______|\_________\   \|__|`,
	`                                 \|_________|        `,
}

// Render renders the full ASCII art Ghost logo using theme colors.
func Render(s *styles.Styles, version string, compact bool, o Opts) string {
	primary := lipgloss.NewStyle().Foreground(s.Primary)

	rendered := make([]string, len(ghostLines))
	for i, line := range ghostLines {
		rendered[i] = primary.Render(line)
	}

	logo := strings.Join(rendered, "\n")

	if version != "" {
		ver := lipgloss.NewStyle().
			Foreground(s.Secondary).
			Faint(true).
			Render("v" + version)
		logo = fmt.Sprintf("%s  %s", logo, ver)
	}

	return logo
}

// SmallRender renders a compact "Ghost" text logo with a gradient and
// trailing diagonal characters that fill the given width.
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

// RenderAnimated renders the logo with a subtle floating animation.
func RenderAnimated(s *styles.Styles, version string, compact bool, o Opts, frame int) string {
	primary := lipgloss.NewStyle().Foreground(s.Primary)

	// Give feeling of floating up/down
	floatOffset := frame % 4
	if floatOffset == 3 {
		floatOffset = 1
	}

	rendered := make([]string, len(ghostLines)+2)
	topPadding := floatOffset
	bottomPadding := 2 - topPadding

	for i := 0; i < topPadding; i++ {
		rendered[i] = "                                                       "
	}

	idx := topPadding
	for _, line := range ghostLines {
		rendered[idx] = primary.Render(line)
		idx++
	}

	for i := 0; i < bottomPadding; i++ {
		rendered[idx] = "                                                       "
		idx++
	}

	logo := strings.Join(rendered, "\n")

	if version != "" {
		ver := lipgloss.NewStyle().
			Foreground(s.Secondary).
			Faint(true).
			Render("v" + version)
		logo = fmt.Sprintf("%s  %s", logo, ver)
	}

	return logo
}
