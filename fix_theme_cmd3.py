import re

with open("internal/cmd/theme.go", "r") as f:
    text = f.read()

text = re.sub(r'// Save the theme.*?m\.saved = true', '// Save the theme\n\t\t\t\tm.currentTheme = m.themes[m.cursor]\n\t\t\t\t_ = m.config.SetTheme(config.ScopeGlobal, m.currentTheme)\n\t\t\t\tm.saved = true', text, flags=re.DOTALL)

with open("internal/cmd/theme.go", "w") as f:
    f.write(text)

