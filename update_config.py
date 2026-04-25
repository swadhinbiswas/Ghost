import re

with open("internal/config/config.go", "r") as f:
    content = f.read()

# Add Theme to TUIOptions definition
new_field = '\tTheme       string `json:"theme,omitempty" jsonschema:"description=TUI coloring theme,default=Default"`\n'
content = re.sub(
    r'(type TUIOptions struct \{\n)',
    r'\1' + new_field,
    content
)

with open("internal/config/config.go", "w") as f:
    f.write(content)

