import re

with open("internal/ui/styles/styles.go", "r") as f:
    text = f.read()

# Remove stray imports at the bottom if any
text = re.sub(r'import "image/color"\nimport "fmt"\n', '', text)

# Remove all hexPtr definitions
text = re.sub(r'func hexPtr\([^)]*\)\s*\*string\s*\{[^}]*\}', '', text, flags=re.DOTALL)

# Add hexPtr at the end
clean_func = """
func hexPtr(c color.Color) *string {
if c == nil {
s := "#000000"
return &s
}
r, g, b, _ := c.RGBA()
s := fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
return &s
}
"""
text = text.strip() + "\n" + clean_func

# Ensure image/color and fmt are imported at the top
if '"image/color"' not in text:
    text = text.replace('import (', 'import (\n\t"image/color"\n\t"fmt"\n', 1)

# Fix bgBaseLighter.Hex issues by using hexPtr
text = re.sub(r'([a-zA-Z0-9_]+)\.Hex', r'hexPtr(\1)', text)

with open("internal/ui/styles/styles.go", "w") as f:
    f.write(text)
