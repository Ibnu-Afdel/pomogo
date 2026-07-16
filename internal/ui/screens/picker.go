package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// DurationPicker renders the Deep Focus duration selection screen.
func DurationPicker(width, height int, th *theme.Theme, selectedIdx int, defaultDuration, workDuration, shortBreakDuration, longBreakDuration time.Duration, sessionsBeforeLongBreak int) string {
	color := lipgloss.Color(th.Accent.String())
	muted := lipgloss.Color(th.Muted.String())
	title := "Deep Focus Duration"
	rhythm := fmt.Sprintf(
		"%s block · %s/%s/%s rhythm every %d",
		formatDuration(defaultDuration),
		formatDuration(workDuration),
		formatDuration(shortBreakDuration),
		formatDuration(longBreakDuration),
		sessionsBeforeLongBreak,
	)

	options := []string{
		"60 min  (1 hour)",
		"120 min (2 hours)",
		"180 min (3 hours)",
		"240 min (4 hours)",
		"Custom duration...",
	}

	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(color).Bold(true).Render(title))
	rows = append(rows, lipgloss.NewStyle().Foreground(muted).Render(rhythm))
	rows = append(rows, "")

	for i, opt := range options {
		indicator := "  "
		itemStyle := lipgloss.NewStyle().Foreground(muted)
		if presetDuration(i) == defaultDuration {
			opt += "  default"
		}
		if i == selectedIdx {
			indicator = "> "
			itemStyle = lipgloss.NewStyle().Foreground(color).Bold(true)
		}
		rows = append(rows, indicator+itemStyle.Render(opt))
	}
	rows = append(rows, "")

	footer := "↓/↑ or tab navigate  ·  1-4 jump  ·  enter select"
	rows = append(rows, lipgloss.NewStyle().Foreground(muted).Render(footer))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(1, 4).
		Width(48).
		Align(lipgloss.Left).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func presetDuration(idx int) time.Duration {
	if idx >= 0 && idx <= 3 {
		return time.Duration(idx+1) * time.Hour
	}
	return 0
}

func formatDuration(d time.Duration) string {
	if d%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(d/time.Hour))
	}
	if d%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(d/time.Minute))
	}
	return d.String()
}
