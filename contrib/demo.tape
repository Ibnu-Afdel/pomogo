# PomoGo demo tape for VHS (https://github.com/charmbracelet/vhs)
# Run with: vhs contrib/demo.tape

Output contrib/demo.gif
Set Theme "Tokyo Night"
Set FontSize 16
Set Width 900
Set Height 480
Set Padding 20

# Launch pomogo
Type "pomogo"
Sleep 500ms
Enter
Sleep 1500ms

# Start a work session
Type "s"
Sleep 3s

# Pause
Type " "
Sleep 2s

# Resume
Type " "
Sleep 2s

# Show help
Type "?"
Sleep 2500ms

# Close help
Type "?"
Sleep 1s

# Skip to break
Type "n"
Sleep 2s

# Quit
Type "q"
Sleep 500ms
