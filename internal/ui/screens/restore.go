package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

// RestorePrompt renders the session restoration prompt screen.
func RestorePrompt(width, height int, th *theme.Theme, activeColor theme.Color) string {
	color := lipgloss.Color(activeColor.String())
	muted := lipgloss.Color(th.Muted.String())

	rows := []string{
		lipgloss.NewStyle().Foreground(color).Bold(true).Render("Restore previous session?"),
		"",
		lipgloss.NewStyle().Foreground(muted).Render("y resume  ·  n discard  ·  q quit"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 4).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
