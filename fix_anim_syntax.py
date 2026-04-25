import re

with open("internal/ui/logo/logo.go", "r") as f:
    text = f.read()

text = text.replace("# just fake blinking", "// just fake blinking")
text = text.replace("line = strings.Replace(line, \"⣿\", \"⠛\", 2) // just fake blinking", "line = strings.Replace(line, \"⣿\", \"⠛\", 2) // just fake blinking")

with open("internal/ui/logo/logo.go", "w") as f:
    f.write(text)

