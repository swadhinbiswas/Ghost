import re

with open("internal/cmd/theme.go", "r") as f:
    text = f.read()

# Fix configState to ConfigStore
text = text.replace('config       *config.ConfigState', 'config       *config.ConfigStore')

# Fix View() string to View() tea.View, and return tea.NewView
text = re.sub(r'func \(m \*themeModel\) View\(\) string \{', 'func (m *themeModel) View() tea.View {', text)
text = re.sub(r'return s\.String\(\)\n\}', 'return tea.NewView(s.String())\n}', text)
text = text.replace('return fmt.Sprintf("\\n  Theme saved as: %s\\n\\n", m.currentTheme)', 'return tea.NewView(fmt.Sprintf("\\n  Theme saved as: %s\\n\\n", m.currentTheme))')


with open("internal/cmd/theme.go", "w") as f:
    f.write(text)

