import re

with open("internal/app/app.go", "r") as f:
    text = f.read()

# Make sure we have lipgloss import
if '"charm.land/lipgloss/v2"' not in text:
    text = text.replace('import (', 'import (\n\t"charm.land/lipgloss/v2"\n', 1)

old_func = """func (a *App) GetTheme() string {
if a.config == nil {
return "Default"
}
cfg := a.config.Config()
if cfg.Options == nil || cfg.Options.TUI == nil || cfg.Options.TUI.Theme == "" {
return "Default"
}
return cfg.Options.TUI.Theme
}"""

new_func = """func (a *App) GetTheme() string {
theme := "Default"
if a.config != nil {
cfg := a.config.Config()
if cfg.Options != nil && cfg.Options.TUI != nil && cfg.Options.TUI.Theme != "" {
theme = cfg.Options.TUI.Theme
}
}

if theme == "Default" || theme == "Auto" {
if lipgloss.HasDarkBackground() {
return "Catppuccin Mocha"
} else {
return "Default"
}
}
return theme
}"""

text = text.replace(old_func, new_func)

with open("internal/app/app.go", "w") as f:
    f.write(text)
