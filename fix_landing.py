import re

with open("internal/ui/model/landing.go", "r") as f:
    text = f.read()

replacement = """func (m *UI) landingView() string {
        t := m.com.Styles
        width := m.layout.main.Dx()
        
        // Let's create a nice frame around the landing info
        
        cwd := common.PrettyPath(t, m.com.Store().WorkingDir(), width)

        parts := []string{
                cwd,
        }

        parts = append(parts, "", m.modelInfo(width))
        infoSection := lipgloss.JoinVertical(lipgloss.Left, parts...)

        _, remainingHeightArea := layout.SplitVertical(m.layout.main, layout.Fixed(lipgloss.Height(infoSection)+1))

        mcpLspSectionWidth := min(40, (width-4)/2)

        lspContent := m.lspInfo(mcpLspSectionWidth-4, max(1, remainingHeightArea.Dy())-4, false)
        mcpContent := m.mcpInfo(mcpLspSectionWidth-4, max(1, remainingHeightArea.Dy())-4, false)

        boxStyle := lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(t.Border).
                Padding(0, 1).
                Width(mcpLspSectionWidth)
                
        lspSection := boxStyle.Render(lspContent)
        mcpSection := boxStyle.Render(mcpContent)

        content := lipgloss.JoinHorizontal(lipgloss.Top, lspSection, "  ", mcpSection)

        return lipgloss.NewStyle().
                Width(width).
                Height(m.layout.main.Dy() - 1).
                PaddingTop(1).
                PaddingLeft(2).
                Render(
                        lipgloss.JoinVertical(lipgloss.Left, infoSection, "", content),
                )
}"""

pattern = r"func \(m \*UI\) landingView\(\) string \{.*?\n\}"
import re
new_text = re.sub(pattern, replacement, text, flags=re.DOTALL)

with open("internal/ui/model/landing.go", "w") as f:
    f.write(new_text)

print("Landing written.")
