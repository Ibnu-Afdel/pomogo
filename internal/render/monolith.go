package render

import (
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	Registry["monolith"] = LayoutSpec{Layout: Monolith, MinWidth: 56, MinHeight: 14}
}

// Monolith renders the timer as a large, nearly standalone object.
func Monolith(ds DisplayState, th *theme.Theme, f Frame) string {
	color := phaseColor(ds.PhaseKind, th)
	muted := lipgloss.Color(th.Muted.String())
	txt := lipgloss.Color(th.Text.String())
	width := f.Width - 8
	if width > 78 {
		width = 78
	}
	if width < 44 {
		width = 44
	}
	center := func(s string) string {
		return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
	}

	var lines []string
	if !ds.Zen {
		lines = append(lines, center(lipgloss.NewStyle().Foreground(muted).Render(strings.ToUpper(ds.ModeLabel))))
		lines = append(lines, "")
	}
	for _, row := range BigClockRows(formatClock(ds), color) {
		lines = append(lines, center(row))
	}
	lines = append(lines, "")
	if ds.Project != "" || ds.Task != "" {
		context := strings.TrimSpace(strings.Join([]string{ds.Project, ds.Task}, "  "))
		lines = append(lines, center(lipgloss.NewStyle().Foreground(txt).Render(TruncateText(context, width, "…"))))
	}
	if !ds.Zen {
		lines = append(lines, center(lipgloss.NewStyle().Foreground(muted).Render(ds.StatusMessage)))
	}
	lines = append(lines, "")
	lines = append(lines, ProgressBar(ds.Progress, width, lipgloss.Color(th.ProgressFill.String()), lipgloss.Color(th.ProgressTrack.String())))

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}
