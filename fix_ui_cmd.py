import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

text = text.replace("m.setState(uiLanding, uiFocusEditor)\n\t\t\t\tcmds = append(cmds, m.tickLanding())", "m.setState(uiLanding, uiFocusEditor)\n\t\t\t\t// cmds = append(cmds, m.tickLanding())")
text = text.replace("m.setState(uiLanding, uiFocusEditor)\n\tcmds = append(cmds, m.tickLanding())", "m.setState(uiLanding, uiFocusEditor)")

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

