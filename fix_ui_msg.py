import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

text = text.replace("type (", "type (\n\ttickLandingMsg time.Time\n")

if "case tickLandingMsg:" not in text:
    old = "case common.EditorCmdMsg:\n\t\t\tcmds = append(cmds, msg())"
    new = "case tickLandingMsg:\n\t\tif m.state == uiLanding {\n\t\t\tm.landingFrame++\n\t\t\tcmds = append(cmds, m.tickLanding())\n\t\t}\n\tcase common.EditorCmdMsg:\n\t\t\tcmds = append(cmds, msg())"
    text = text.replace(old, new)


with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

