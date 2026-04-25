import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

# Let's ensure tickLanding gets called!
text = text.replace("m.landingFrame = 0", "m.landingFrame = 0\n\t\t\t\tcmds = append(cmds, m.tickLanding())")

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

