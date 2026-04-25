import re

with open("internal/ui/logo/logo.go", "r") as f:
    text = f.read()

text = text.replace("strings.Join(rendered, \"\\n\")", "strings.Join(rendered, \"\\\\n\")")
text = text.replace('strings.Join(rendered, "\n")', 'strings.Join(rendered, "\\n")')

with open("internal/ui/logo/logo.go", "w") as f:
    f.write(text)

