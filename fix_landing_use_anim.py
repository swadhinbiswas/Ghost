import re

with open("internal/ui/model/landing.go", "r") as f:
    text = f.read()

# Make sure logo is imported
if '"github.com/swadhinbiswas/ghost/internal/ui/logo"' not in text:
    text = text.replace('"github.com/swadhinbiswas/ghost/internal/ui/common"', '"github.com/swadhinbiswas/ghost/internal/ui/common"\n\t"github.com/swadhinbiswas/ghost/internal/ui/logo"')

# Replace Welcome text with animated Ghost logo
old_title = 'title := t.Base.Foreground(t.Primary).Bold(true).Render("✨ Welcome to Ghost ✨")'
new_title = 'title := logo.RenderAnimated(t, "", false, logo.Opts{}, m.landingFrame)'
text = text.replace(old_title, new_title)

with open("internal/ui/model/landing.go", "w") as f:
    f.write(text)

