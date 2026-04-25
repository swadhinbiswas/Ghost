import re

with open("internal/ui/styles/styles.go", "r") as f:
    text = f.read()

# Fix fgMuted without hexPtr in ImageText
text = text.replace('Color:  fgMuted,', 'Color:  hexPtr(fgMuted),')
text = text.replace('Color: fgMuted,', 'Color: hexPtr(fgMuted),')

# Fix plainBg and plainFg
text = text.replace('plainBg := new(bgBaseLighter())', 'plainBg := hexPtr(bgBaseLighter)')
text = text.replace('plainFg := new(fgMuted())', 'plainFg := hexPtr(fgMuted)')
text = text.replace('plainBg := new(bgBaseLighter)', 'plainBg := hexPtr(bgBaseLighter)')
text = text.replace('plainFg := new(fgMuted)', 'plainFg := hexPtr(fgMuted)')

with open("internal/ui/styles/styles.go", "w") as f:
    f.write(text)
