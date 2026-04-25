import re

with open("internal/cmd/session.go", "r") as f:
    text = f.read()

# Make sure we have lipgloss import
if '"charm.land/lipgloss/v2"' not in text:
    text = text.replace('import (', 'import (\n\t"charm.land/lipgloss/v2"\n', 1)

text = re.sub(r' sty := styles\.DefaultStyles\(themeName\)', r"""
if themeName == "Default" || themeName == "Auto" {
if lipgloss.HasDarkBackground() {
themeName = "Catppuccin Mocha"
} else {
themeName = "Default"
}
}
sty := styles.DefaultStyles(themeName)
""", text, count=1)

with open("internal/cmd/session.go", "w") as f:
    f.write(text)
