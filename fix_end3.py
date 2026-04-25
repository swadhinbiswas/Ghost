import re

with open("internal/ui/styles/styles.go", "r") as f:
    text = f.read()

# fgMuted and bgBaseLighter need to be passed to hexPtr to get a *string.
# Let's fix lines 713 and 808
text = text.replace('Color: fgMuted', 'Color: hexPtr(fgMuted)')
text = text.replace('Background: bgBaseLighter', 'Background: hexPtr(bgBaseLighter)')

# Line 826 and 827 have `bgBaseLighter(bgBaseLighter)` which is broken.
text = re.sub(r'bgBaseLighter\(bgBaseLighter\)', 'hexPtr(bgBaseLighter)', text)
text = re.sub(r'fgMuted\(fgMuted\)', 'hexPtr(fgMuted)', text)

with open("internal/ui/styles/styles.go", "w") as f:
    f.write(text)
