import re

with open("internal/ui/model/landing.go", "r") as f:
    text = f.read()

text = text.replace("""// landingView renders the landing page view showing the current working
// directory, model information, and LSP/MCP status in a two-column layout.""", """// landingView renders the landing page view showing the current working
// directory and keyboard shortcuts.""")

old_info = """        infoSection := lipgloss.JoinVertical(lipgloss.Center, 
                title,
                "",
                cwd,
                "",
                m.modelInfo(width),
        )"""

new_info = """        infoSection := lipgloss.JoinVertical(lipgloss.Center, 
                title,
                "",
                cwd,
        )"""

text = text.replace(old_info, new_info)

with open("internal/ui/model/landing.go", "w") as f:
    f.write(text)

