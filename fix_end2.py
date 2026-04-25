import re

with open("internal/ui/styles/styles.go", "r") as f:
    text = f.read()

# Make sure fmt is imported
if '"fmt"' not in text:
    text = text.replace('import (', 'import (\n\t"fmt"\n', 1)

# Fix hexPtr() misusage, bgBaseLighter and fgMuted are ALREADY hexPtr strings in `styles.go` now, let's look at how they are defined.
# If they are strings already, we shouldn't cast them... 
text = text.replace('hexPtr(bgBaseLighter)', 'bgBaseLighter')
text = text.replace('hexPtr(fgMuted)', 'fgMuted')

with open("internal/ui/styles/styles.go", "w") as f:
    f.write(text)
