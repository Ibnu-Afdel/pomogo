package screens

import (
	"fmt"
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/notify"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// SoundPicker renders the sound profile selection screen.
func SoundPicker(width, height int, th *theme.Theme, selectedIdx int, profiles []notify.SoundProfile) string {
	color := lipgloss.Color(th.Accent.String())
	muted := lipgloss.Color(th.Muted.String())
	txt := lipgloss.Color(th.Text.String())

	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(color).Bold(true).Render("Sound Profile"))
	rows = append(rows, lipgloss.NewStyle().Foreground(muted).Render("focus start / focus end"))
	rows = append(rows, "")

	for i, profile := range profiles {
		indicator := "  "
		nameStyle := lipgloss.NewStyle().Foreground(txt)
		if i == selectedIdx {
			indicator = "> "
			nameStyle = lipgloss.NewStyle().Foreground(color).Bold(true)
		}
		line := fmt.Sprintf("%-8s  %s / %s", profile.Name, profile.StartEvent, profile.EndEvent)
		rows = append(rows, indicator+nameStyle.Render(line))
		rows = append(rows, "  "+lipgloss.NewStyle().Foreground(muted).Render(profile.Description))
	}

	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(muted).Render("↓/↑ or tab navigate  ·  space preview  ·  enter select"))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 4).
		Width(64).
		Align(lipgloss.Left).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
