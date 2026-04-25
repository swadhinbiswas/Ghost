import re

with open("internal/ui/model/ui.go", "r") as f:
    text = f.read()

# Add tick msg type
text = re.sub(r'type \(\s+clearScreenMsg struct\{\}', r'type (\n\ttickLandingMsg time.Time\n\n\tclearScreenMsg struct{}', text)

# Add tick handler in Update
tick_handler = """case tickLandingMsg:
        if m.state == uiLanding {
                m.landingFrame++
                cmds = append(cmds, m.tickLanding())
        }
"""
text = re.sub(r'(case common\.EditorCmdMsg:\s+cmds = append\(cmds, msg\(\)\)\s+)', r'\1' + tick_handler, text)

with open("internal/ui/model/ui.go", "w") as f:
    f.write(text)

with open("internal/ui/model/landing.go", "a") as f:
    f.write("\nfunc (m *UI) tickLanding() tea.Cmd {\n\treturn tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg {\n\t\treturn tickLandingMsg(t)\n\t})\n}\n")
