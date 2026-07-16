package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

// DurationPicker renders the Deep Focus duration selection screen.
func DurationPicker(width, height int, th *theme.Theme, selectedIdx int) string {
	color := lipgloss.Color(th.Accent.String())
	title := "Deep Focus Duration"

	options := []string{
		"1 hour",
		"2 hours",
		"3 hours",
		"4 hours",
		"Custom duration...",
	}

	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(color).Bold(true).Render(title))
	rows = append(rows, "")

	for i, opt := range options {
		indicator := "  "
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted.String()))
		if i == selectedIdx {
			indicator = "> "
			itemStyle = lipgloss.NewStyle().Foreground(color).Bold(true)
		}
		rows = append(rows, indicator+itemStyle.Render(opt))
	}
	rows = append(rows, "")

	footer := "↓/↑ navigate  ·  enter select  ·  esc cancel"
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted.String())).Render(footer))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 4).
		Width(40).
		Align(lipgloss.Left).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
