import re

with open("internal/cmd/session.go", "r") as f:
    text = f.read()

# Add a little code to outputSessionHuman to load config. We have the generic config loader logic we can just inject.
# Actually, since it's just human output, it's not super critical, but we should do it properly.

snippet = """
themeName := "Default"
cwd, _ := os.Getwd()
if cfg, err := config.Load(cwd, "", false); err == nil {
if cfg.Config().Options != nil && cfg.Config().Options.TUI != nil && cfg.Config().Options.TUI.Theme != "" {
themeName = cfg.Config().Options.TUI.Theme
}
}
sty := styles.DefaultStyles(themeName)
"""

text = text.replace('sty := styles.DefaultStyles("Default")', snippet)

with open("internal/cmd/session.go", "w") as f:
    f.write(text)
