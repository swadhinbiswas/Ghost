import re

with open("internal/app/app.go", "r") as f:
    content = f.read()

# I want to add a function to app to get the theme safely
helper = """

func (a *App) GetTheme() string {
if a.config == nil {
return "Default"
}
cfg := a.config.Config()
if cfg.Options == nil || cfg.Options.TUI == nil || cfg.Options.TUI.Theme == "" {
return "Default"
}
return cfg.Options.TUI.Theme
}
"""

if "func (a *App) GetTheme" not in content:
    content = content + helper

# Use app.GetTheme()
content = content.replace('styles.DefaultStyles("Default")', 'styles.DefaultStyles(app.GetTheme())')

with open("internal/app/app.go", "w") as f:
    f.write(content)
