import re

with open("internal/ui/logo/logo.go", "r") as f:
    text = f.read()

replacement = """func RenderAnimated(s *styles.Styles, version string, compact bool, o Opts, frame int) string {
        purple := lipgloss.NewStyle().Foreground(colorPurple)
        pink   := lipgloss.NewStyle().Foreground(colorPink)
        fg     := lipgloss.NewStyle().Foreground(colorFG)

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
                if i == 3 && frame % 10 < 2 {
                    line = line.replace("⣿", "⠛", 2) # just fake blinking
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
"""

with open("internal/ui/logo/logo.go", "a") as f:
    f.write("\n" + replacement)
print("added")
