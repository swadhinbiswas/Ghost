import re

with open("internal/ui/model/landing.go", "r") as f:
    text = f.read()

replacement = """func (m *UI) landingView() string {
        t := m.com.Styles
        width := m.layout.main.Dx()
        height := m.layout.main.Dy() - 1
        
        cwd := common.PrettyPath(t, m.com.Store().WorkingDir(), width)

        title := t.Base.Foreground(t.Primary).Bold(true).Render("✨ Welcome to Ghost ✨")
        
        infoSection := lipgloss.JoinVertical(lipgloss.Center, 
                title,
                "",
                cwd,
                "",
                m.modelInfo(width),
        )
        
        infoSection = lipgloss.PlaceHorizontal(width, lipgloss.Center, infoSection)

        mcpLspSectionWidth := min(46, (width-10)/2)
        mcpLspSectionWidth = max(25, mcpLspSectionWidth)

        lspContent := m.lspInfo(mcpLspSectionWidth-4, 5, false)
        mcpContent := m.mcpInfo(mcpLspSectionWidth-4, 5, false)

        boxStyle := lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                BorderForeground(t.Border).
                Padding(1, 2).
                Width(mcpLspSectionWidth).
                Height(6)
                
        lspSection := boxStyle.Render(lspContent)
        mcpSection := boxStyle.Render(mcpContent)

        content := lipgloss.JoinHorizontal(lipgloss.Top, lspSection, "    ", mcpSection)
        content = lipgloss.PlaceHorizontal(width, lipgloss.Center, content)

        finalView := lipgloss.JoinVertical(lipgloss.Center, infoSection, "", "", content)
        
        return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, finalView)
}"""

pattern = r"func \(m \*UI\) landingView\(\) string \{.*?\n\}"
import re
new_text = re.sub(pattern, replacement, text, flags=re.DOTALL)

with open("internal/ui/model/landing.go", "w") as f:
    f.write(new_text)

print("Landing written.")
