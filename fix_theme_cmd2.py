import re

with open("internal/cmd/theme.go", "r") as f:
    text = f.read()

# Replace block where we save config
block = """
// Save the theme
m.currentTheme = m.themes[m.cursor]
c := m.config.Config()
if c.Options == nil {
c.Options = &config.Options{}
}
if c.Options.TUI == nil {
c.Options.TUI = &config.TUIOptions{}
}
c.Options.TUI.Theme = m.currentTheme
// save config
err := m.config.Save()
if err != nil {
// handle err visually if needed, but for now just quit
}
"""

replacement = """
// Save the theme
m.currentTheme = m.themes[m.cursor]
_ = m.config.SetTheme(config.ScopeGlobal, m.currentTheme)
"""

text = text.replace(block, replacement)

with open("internal/cmd/theme.go", "w") as f:
    f.write(text)

