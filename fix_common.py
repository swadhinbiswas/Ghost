import re

with open("internal/ui/common/common.go", "r") as f:
    content = f.read()

content = content.replace('styles.DefaultStyles("Default")', 'styles.DefaultStyles(app.GetTheme())')

with open("internal/ui/common/common.go", "w") as f:
    f.write(content)
