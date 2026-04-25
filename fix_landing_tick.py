import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

# Add tick handler in Update
text = re.sub(r'(m\.setState\(uiLanding, uiFocusEditor\))', r'\1\n\t\t\t\tcmds = append(cmds, m.tickLanding())', text)

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

