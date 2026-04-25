import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

replacement = """                                if value == "exit" || value == "quit" {
                                        return m.openQuitDialog()
                                }

                                if strings.HasPrefix(value, "/theme ") {
                                        themeName := strings.TrimSpace(strings.TrimPrefix(value, "/theme "))
                                        if _, ok := styles.ThemeMap[themeName]; ok {
                                                err := m.com.Store().SetTheme(config.ScopeGlobal, themeName)
                                                if err != nil {
                                                        return util.ReportError(err)
                                                }
                                                // Update locally
                                                m.com.Styles = styles.DefaultStyles(themeName)
                                                // Request a redraw
                                                cmds = append(cmds, util.ReportInfo("Switched theme to " + themeName))
                                                return tea.Batch(cmds...)
                                        } else {
                                                return util.ReportWarn("Theme not found: " + themeName)
                                        }
                                }"""

text = text.replace(
"""                                if value == "exit" || value == "quit" {
                                        return m.openQuitDialog()
                                }""",
replacement
)

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

print("done")
