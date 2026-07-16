package devinfo

import (
	"os"
	"os/exec"
	"strings"
)

// GetTmuxSession queries and returns the tmux session name if inside TMUX, or empty string.
func GetTmuxSession() string {
	if os.Getenv("TMUX") == "" {
		return ""
	}
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
