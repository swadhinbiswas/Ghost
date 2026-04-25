import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

text = text.replace("styles.ThemeMap", "styles.Themes")

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

