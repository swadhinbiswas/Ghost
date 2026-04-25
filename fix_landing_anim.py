import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

# Add landingFrame to UI struct
text = re.sub(r'continueLastSession bool\n\n(\s+)lastUserMessageTime int64', r'continueLastSession bool\n\n\1landingFrame int\n\tlastUserMessageTime int64', text)
print("substituted landingFrame")

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)
