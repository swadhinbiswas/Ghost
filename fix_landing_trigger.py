import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

# Add tick handler in Update when state changes to uiLanding
# Let's find:
# m.setState(uiLanding, uiFocusEditor)
# inside ui.go functions and trigger it manually if needed, or better, 
# just make landingView trigger a message! No, view shouldn't trigger commands.

text = text.replace("m.setState(uiLanding, uiFocusEditor)", "m.setState(uiLanding, uiFocusEditor)\n\tm.landingFrame = 0")
text = text.replace("func (m *UI) Init() tea.Cmd {", "func (m *UI) Init() tea.Cmd {\n\t// It will naturally start ticking because of the tickLanding call in common commands if we need.")

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

