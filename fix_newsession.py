import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

text = text.replace("m.landingFrame = 0\n        cmds = append(cmds, m.tickLanding())", "m.landingFrame = 0\n")

# Wait, `tea.Batch( ... )`
batch_old = "        return tea.Batch(\n                func() tea.Msg {\n                        m.com.App.Permissions.ClearSessionPerms()\n                        return nil\n                },"
batch_new = "        return tea.Batch(m.tickLanding(),\n                func() tea.Msg {\n                        m.com.App.Permissions.ClearSessionPerms()\n                        return nil\n                },"
text = text.replace(batch_old, batch_new)

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

