package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

// Input renders the text input overlay screen (for task, note, or project inputs).
func Input(width, height int, th *theme.Theme, mode string, inputView string, suggestions []string, suggestionIndex int) string {
	var color lipgloss.Color
	var title string

	switch mode {
	case "note":
		color = lipgloss.Color(th.Work.String())
		title = "Session Note"
	case "project":
		color = lipgloss.Color(th.Accent.String())
		title = "Set Project Name"
	case "custom_duration":
		color = lipgloss.Color(th.Accent.String())
		title = "Set Custom Duration (e.g. 1h30m, 90m)"
	default: // "task"
		color = lipgloss.Color(th.Accent.String())
		title = "Set Current Task"
	}

	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(color).Bold(true).Render(title))
	rows = append(rows, "")
	rows = append(rows, inputView)
	rows = append(rows, "")

	if (mode == "task" || mode == "project") && len(suggestions) > 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(color).Bold(true).Render("Suggestions:"))
		for i, s := range suggestions {
			if i >= 5 {
				break
			}
			indicator := "  "
			itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted.String()))
			if i == suggestionIndex {
				indicator = "> "
				itemStyle = lipgloss.NewStyle().Foreground(color).Bold(true)
			}
			rows = append(rows, indicator+itemStyle.Render(s))
		}
		rows = append(rows, "")
	}

	footer := "enter save  ·  esc cancel"
	if (mode == "task" || mode == "project") && len(suggestions) > 0 {
		footer = "↓/↑ navigate  ·  tab select  ·  ctrl+d delete  ·  enter save"
	}
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
