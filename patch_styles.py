import re

with open('internal/ui/styles/styles.go', 'r') as f:
    content = f.read()

# Make DefaultStyles take a theme string
content = content.replace('func DefaultStyles() Styles {', 'func DefaultStyles(themeName string) Styles {')

replacement = "t := GetTheme(themeName)"
	var (
primary       = t.Primary
secondary     = t.Secondary
tertiary      = t.Tertiary
bgBase        = t.BgBase
bgBaseLighter = t.BgBaseLighter
bgSubtle      = t.BgSubtle
bgOverlay     = t.BgOverlay
fgBase        = t.FgBase
fgMuted       = t.FgMuted
fgHalfMuted   = t.FgHalfMuted
fgSubtle      = t.FgSubtle
border        = t.Border
borderFocus   = t.BorderFocus
error         = t.Error
warning       = t.Warning
info          = t.Info
white         = t.White
blueLight     = t.BlueLight
blue          = t.Blue
blueDark      = t.BlueDark
yellow        = t.Yellow
greenLight    = t.GreenLight
green         = t.Green
greenDark     = t.GreenDark
red           = t.Red
redDark       = t.RedDark
)"""

# We need to replace the charmtone assignments block
pattern = r'var \(\n\s*primary\s*=\s*charmtone\.Charple.*?cherry\s*=\s*charmtone\.Cherry\n\s*\)'

content = re.sub(pattern, replacement, content, flags=re.DOTALL)

helper_func = """
func hexPtr(c lipgloss.TerminalColor) *string {
	if col, ok := c.(lipgloss.Color); ok {
		s := string(col)
		return &s
	}
	s := "#000000"
	return &s
}
"""
content = content.replace('func chromaStyle(style ansi.StylePrimitive) string {', helper_func + '\nfunc chromaStyle(style ansi.StylePrimitive) string {')

content = content.replace('new(charmtone.Smoke.Hex())', 'hexPtr(fgHalfMuted)')
content = content.replace('new(charmtone.Malibu.Hex())', 'hexPtr(blue)')
content = content.replace('new(charmtone.Zest.Hex())', 'hexPtr(warning)')
content = content.replace('new(charmtone.Charple.Hex())', 'hexPtr(primary)')
content = content.replace('new(charmtone.Guac.Hex())', 'hexPtr(greenDark)')
content = content.replace('new(charmtone.Charcoal.Hex())', 'hexPtr(bgSubtle)')
content = content.replace('new(charmtone.Zinc.Hex())', 'hexPtr(blueLight)')
content = content.replace('new(charmtone.Cheeky.Hex())', 'hexPtr(secondary)')
content = content.replace('new(charmtone.Squid.Hex())', 'hexPtr(fgMuted)')
content = content.replace('new(charmtone.Coral.Hex())', 'hexPtr(red)')
content = content.replace('new(charmtone.Butter.Hex())', 'hexPtr(white)')
content = content.replace('new(charmtone.Sriracha.Hex())', 'hexPtr(redDark)')
content = content.replace('new(charmtone.Oyster.Hex())', 'hexPtr(fgSubtle)')
content = content.replace('new(charmtone.Bengal.Hex())', 'hexPtr(yellow)')
content = content.replace('new(charmtone.Pony.Hex())', 'hexPtr(primary)')
content = content.replace('new(charmtone.Guppy.Hex())', 'hexPtr(greenLight)')
content = content.replace('new(charmtone.Salmon.Hex())', 'hexPtr(red)')
content = content.replace('new(charmtone.Mauve.Hex())', 'hexPtr(primary)')
content = content.replace('new(charmtone.Hazy.Hex())', 'hexPtr(blueLight)')
content = content.replace('new(charmtone.Salt.Hex())', 'hexPtr(white)')
content = content.replace('new(charmtone.Citron.Hex())', 'hexPtr(yellow)')
content = content.replace('new(charmtone.Julep.Hex())', 'hexPtr(green)')
content = content.replace('new(charmtone.Cumin.Hex())', 'hexPtr(secondary)')
content = content.replace('new(charmtone.Bok.Hex())', 'hexPtr(tertiary)')

# Fix standalone charmtone. references in styles
content = content.replace('charmtone.Oyster', 'fgSubtle')
content = content.replace('charmtone.Citron', 'yellow')
content = content.replace('charmtone.Pepper', 'bgBase')
content = content.replace('charmtone.Squid', 'fgMuted')
content = content.replace('charmtone.Zest', 'warning')
content = content.replace('charmtone.Charcoal', 'bgSubtle')
content = content.replace('charmtone.Iron', 'bgOverlay')
content = content.replace('charmtone.Coral', 'red')
content = content.replace('charmtone.Guac', 'greenDark')
content = content.replace('charmtone.Salt', 'white')
content = content.replace('charmtone.Charple', 'primary')
content = content.replace('charmtone.Butter', 'white')
content = content.replace('charmtone.Bok', 'tertiary')

with open('internal/ui/styles/styles.go', 'w') as f:
    f.write(content)
