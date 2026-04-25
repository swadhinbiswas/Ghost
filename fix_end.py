import re

with open("internal/ui/styles/styles.go", "r") as f:
    text = f.read()

# Remove everything after the Render function's closing brace
idx = text.rfind('return s.String()\n}')
if idx != -1:
    text = text[:idx + len('return s.String()\n}')]

func_str = """
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
text = text + "\n\n" + func_str

# Also fix the weird lipgloss.Color typecast bug that might still be there:
# Wait, let's just make sure we drop lipgloss.Color cast where it doesn't belong
text = text.replace('lipgloss.Color(fgMuted)', 'fgMuted')
text = text.replace('lipgloss.Color(bgBaseLighter)', 'bgBaseLighter')

with open("internal/ui/styles/styles.go", "w") as f:
    f.write(text)
