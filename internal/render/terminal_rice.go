package render

import (
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	Registry["terminal-rice"] = LayoutSpec{Layout: TerminalRice, MinWidth: 56, MinHeight: 16}
}

// TerminalRice renders a styled terminal-rice frame while keeping density restrained.
func TerminalRice(ds DisplayState, th *theme.Theme, f Frame) string {
	color := phaseColor(ds.PhaseKind, th)
	muted := lipgloss.Color(th.Muted.String())
	border := lipgloss.Color(th.Border.String())
	txt := lipgloss.Color(th.Text.String())
	width := f.Width - 10
	if width > 72 {
		width = 72
	}
	if width < 46 {
		width = 46
	}
	center := func(s string) string {
		return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
	}

	rail := lipgloss.NewStyle().Foreground(border).Render(strings.Repeat("─", width))
	header := center(lipgloss.NewStyle().Foreground(muted).Render("✦ " + strings.ToUpper(ds.ModeLabel) + " ✦"))
	var body []string
	body = append(body, rail, header, "")
	for _, row := range BigClockRows(formatClock(ds), color) {
		body = append(body, center(row))
	}
	body = append(body, "")
	if ds.Project != "" {
		body = append(body, center(lipgloss.NewStyle().Foreground(txt).Bold(true).Render(TruncateText(ds.Project, width, "…"))))
	}
	if ds.Task != "" && !ds.Zen {
		body = append(body, center(lipgloss.NewStyle().Foreground(muted).Render(TruncateText(ds.Task, width, "…"))))
	}
	body = append(body, "", ProgressBar(ds.Progress, width, lipgloss.Color(th.ProgressFill.String()), lipgloss.Color(th.ProgressTrack.String())), rail)

	box := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(border).
		Padding(0, 2).
		Render(strings.Join(body, "\n"))
	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, box)
}
