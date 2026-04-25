with open("internal/ui/styles/theme.go", "r") as f:
    lines = f.readlines()

new_lines = []
skip = False
for line in lines:
    if line.strip() == '"Tokyo Night": {' and new_lines and '"Tokyo Night":' in "".join(new_lines[-100:]):
        # Already have Tokyo Night, we probably have the others duplicated too
        pass
    new_lines.append(line)

# Let's do a better way: parse and rewrite
