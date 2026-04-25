import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

replacement = """                                if strings.HasPrefix(value, "/theme") {
                                        args := strings.TrimSpace(strings.TrimPrefix(value, "/theme"))

                                        if args == "" {
                                            var names []string
                                            for name := range styles.Themes {
                                                names = append(names, name)
                                            }
                                            return util.ReportInfo("Available themes: " + strings.Join(names, ", "))()
                                        }

                                        if _, ok := styles.Themes[args]; ok {
                                                err := m.com.Store().SetTheme(config.ScopeGlobal, args)
                                                if err != nil {
                                                        return util.ReportError(err)()
                                                }
                                                // Update locally
                                                m.com.Styles = styles.DefaultStyles(args)
                                                // Request a redraw
                                                cmds = append(cmds, util.ReportInfo("Switched theme to " + args)())
                                                return tea.Batch(cmds...)
                                        } else {
                                                return util.ReportWarn("Theme not found. Try /theme to list them.")()
                                        }
                                }"""

# Using regex to replace the old /theme block
pattern = r"""\s+if strings\.HasPrefix\(value, "/theme "\) \{.*?return util\.ReportWarn\("Theme not found: " \+ themeName\).*?\}"""
text = re.sub(pattern, replacement, text, flags=re.DOTALL)

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

