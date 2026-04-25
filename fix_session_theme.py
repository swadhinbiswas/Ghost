import re

with open("internal/cmd/session.go", "r") as f:
    text = f.read()

# Make sure we have lipgloss import
if '"charm.land/lipgloss/v2"' not in text:
    text = text.replace('import (', 'import (\n\t"charm.land/lipgloss/v2"\n', 1)

old_logic = "if cfg, err := config.Load(cwd, "", false); err == nil {"
		if cfg.Config().Options != nil && cfg.Config().Options.TUI != nil && cfg.Config().Options.TUI.Theme != "" {
			themeName = cfg.Config().Options.TUI.Theme
		}
	}
	sty := styles.DefaultStyles(themeName)"""

new_logic = """if cfg, err := config.Load(cwd, "", false); err == nil {
		if cfg.Config().Options != nil && cfg.Config().Options.TUI != nil && cfg.Config().Options.TUI.Theme != "" {
			themeName = cfg.Config().Options.TUI.Theme
		}
	}
	if themeName == "Default" || themeName == "Auto" {
		if lipgloss.HasDarkBackground() {
			themeName = "Catppuccin Mocha"
		} else {
			themeName = "Default"
		}
	}
	sty := styles.DefaultStyles(themeName)"""

text = text.replace(old_logic, new_logic)

with open("internal/cmd/session.go", "w") as f:
    f.write(text)
