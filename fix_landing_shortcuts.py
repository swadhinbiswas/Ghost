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
}"""

pattern = r"func \(m \*UI\) landingView\(\) string \{.*?\n\}"
import re
new_text = re.sub(pattern, replacement, text, flags=re.DOTALL)

with open("internal/ui/model/landing.go", "w") as f:
    f.write(new_text)

print("Landing shortcuts written.")
