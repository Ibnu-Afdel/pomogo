package render

import (
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	Registry["focus-stack"] = LayoutSpec{Layout: FocusStack, MinWidth: 54, MinHeight: 14}
}

// FocusStack renders a compact stacked layout for repeated work sessions.
func FocusStack(ds DisplayState, th *theme.Theme, f Frame) string {
	color := phaseColor(ds.PhaseKind, th)
	muted := lipgloss.Color(th.Muted.String())
	txt := lipgloss.Color(th.Text.String())
	border := lipgloss.Color(th.Border.String())
	width := f.Width - 8
	if width > 72 {
		width = 72
	}
	if width < 44 {
		width = 44
	}

	left := lipgloss.NewStyle().Foreground(color).Bold(true).Render(formatClock(ds))
	right := lipgloss.NewStyle().Foreground(muted).Render(strings.ToUpper(ds.ModeLabel))
	headerGap := width - TextWidth(formatClock(ds)) - TextWidth(ds.ModeLabel)
	if headerGap < 1 {
		headerGap = 1
	}
	header := left + strings.Repeat(" ", headerGap) + right

	context := strings.TrimSpace(strings.Join([]string{ds.Project, ds.Task}, "  /  "))
	if context == "" {
		context = ds.StatusMessage
	}
	context = TruncateText(context, width, "…")

	bar := ProgressBar(ds.Progress, width, lipgloss.Color(th.ProgressFill.String()), lipgloss.Color(th.ProgressTrack.String()))
	rail := lipgloss.NewStyle().Foreground(border).Render(strings.Repeat("─", width))
	lines := []string{
		header,
		rail,
		lipgloss.NewStyle().Foreground(txt).Render(context),
		lipgloss.NewStyle().Foreground(muted).Render(TruncateText(ds.StatusMessage, width, "…")),
		"",
		bar,
	}

	if ds.Zen {
		lines = []string{header, rail, lipgloss.NewStyle().Foreground(txt).Render(context), "", bar}
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Render(strings.Join(lines, "\n"))
	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, box)
}
