import re

with open("internal/config/config.go", "r") as f:
    config_text = f.read()

# remove extra Theme fields
parts = config_text.split('\tTheme       string `json:"theme,omitempty" jsonschema:"description=TUI coloring theme,default=Default"`\n')
# keep only the first one and the rest
if len(parts) > 2:
    config_text = parts[0] + '\tTheme       string `json:"theme,omitempty" jsonschema:"description=TUI coloring theme,default=Default"`\n' + "".join(parts[1:]).replace('\tTheme       string `json:"theme,omitempty" jsonschema:"description=TUI coloring theme,default=Default"`\n', '')
with open("internal/config/config.go", "w") as f:
    f.write(config_text)

with open("internal/ui/styles/theme.go", "r") as f:
    theme_text = f.read()

# remove everything between 'var Themes = map[string]Theme{' and '}'
match = re.search(r'var Themes = map\[string\]Theme\{.*?^\s*\}\s*$', theme_text, re.MULTILINE | re.DOTALL)
if match:
    # Just recreate the map to be safe
    pass

