import re
with open("internal/ui/styles/styles.go", "r") as f:
    content = f.read()

# remove all occurrences of hexPtr function
content = re.sub(r'func hexPtr.*?return &s\n}', '', content, flags=re.DOTALL)

# add the correct hexPtr at the end
clean_func = """
import "image/color"
import "fmt"

func hexPtr(c color.Color) *string {
r, g, b, _ := c.RGBA()
s := fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
return &s
}
"""
content += clean_func

with open("internal/ui/styles/styles.go", "w") as f:
    f.write(content)

