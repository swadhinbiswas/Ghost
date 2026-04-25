import re

with open('internal/ui/styles/styles.go', 'r') as f:
    content = f.read()

content = content.replace('func DefaultStyles() Styles {', 'func DefaultStyles(themeName string) Styles {')

# The exact string mapping we need to target is:
pattern = r'var \(\n\s*primary\s*=\s*charmtone\.Charple.*?cherry\s*=\s*charmtone\.Cherry\n\s*\)'
replacement = """\tt := GetTheme(themeName)
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

content = re.sub(pattern, replacement, content, flags=re.DOTALL)

# Add hex helper at the very bottom
content += """

func hexPtr(c any) *string {
if col, ok := c.(lipgloss.Color); ok {
s := string(col)
return &s
}
s := "#000000"
return &s
}
"""

color_map = {
    'charmtone.Smoke.Hex()': 'hexPtr(fgHalfMuted)',
    'charmtone.Malibu.Hex()': 'hexPtr(blue)',
    'charmtone.Zest.Hex()': 'hexPtr(warning)',
    'charmtone.Charple.Hex()': 'hexPtr(primary)',
    'charmtone.Guac.Hex()': 'hexPtr(greenDark)',
    'charmtone.Charcoal.Hex()': 'hexPtr(bgSubtle)',
    'charmtone.Zinc.Hex()': 'hexPtr(blueLight)',
    'charmtone.Cheeky.Hex()': 'hexPtr(secondary)',
    'charmtone.Squid.Hex()': 'hexPtr(fgMuted)',
    'charmtone.Coral.Hex()': 'hexPtr(red)',
    'charmtone.Butter.Hex()': 'hexPtr(white)',
    'charmtone.Sriracha.Hex()': 'hexPtr(redDark)',
    'charmtone.Oyster.Hex()': 'hexPtr(fgSubtle)',
    'charmtone.Bengal.Hex()': 'hexPtr(yellow)',
    'charmtone.Pony.Hex()': 'hexPtr(primary)',
    'charmtone.Guppy.Hex()': 'hexPtr(greenLight)',
    'charmtone.Salmon.Hex()': 'hexPtr(red)',
    'charmtone.Mauve.Hex()': 'hexPtr(primary)',
    'charmtone.Hazy.Hex()': 'hexPtr(blueLight)',
    'charmtone.Salt.Hex()': 'hexPtr(white)',
    'charmtone.Citron.Hex()': 'hexPtr(yellow)',
    'charmtone.Julep.Hex()': 'hexPtr(green)',
    'charmtone.Cumin.Hex()': 'hexPtr(secondary)',
    'charmtone.Bok.Hex()': 'hexPtr(tertiary)'
}

for k, v in color_map.items():
    content = content.replace(f'new({k})', v)
    content = content.replace(f'*{k}', f'*{v}')

lit_map = {
    'charmtone.Oyster': 'fgSubtle',
    'charmtone.Citron': 'yellow',
    'charmtone.Pepper': 'bgBase',
    'charmtone.Squid': 'fgMuted',
    'charmtone.Zest': 'warning',
    'charmtone.Charcoal': 'bgSubtle',
    'charmtone.Iron': 'bgOverlay',
    'charmtone.Coral': 'red',
    'charmtone.Guac': 'greenDark',
    'charmtone.Salt': 'white',
    'charmtone.Charple': 'primary',
    'charmtone.Butter': 'white',
    'charmtone.Bok': 'tertiary',
    'charmtone.Damson': 'blueDark',
    'charmtone.Mustard': 'yellow',
    'charmtone.Sardine': 'blueLight'
}

for k, v in lit_map.items():
    content = content.replace(k, v)

content = content.replace('"github.com/charmbracelet/x/exp/charmtone"', '')

with open('internal/ui/styles/styles.go', 'w') as f:
    f.write(content)

