package render

import (
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	Registry["tinybar"] = LayoutSpec{Layout: Tinybar, MinWidth: 40, MinHeight: 4}
}

// Tinybar renders a narrow status strip for terminal edges and small tmux panes.
func Tinybar(ds DisplayState, th *theme.Theme, f Frame) string {
	color := phaseColor(ds.PhaseKind, th)
	muted := lipgloss.Color(th.Muted.String())
	txt := lipgloss.Color(th.Text.String())
	width := f.Width - 4
	if width > 90 {
		width = 90
	}
	if width < 36 {
		width = 36
	}

	left := lipgloss.NewStyle().Foreground(color).Bold(true).Render(formatClock(ds))
	label := lipgloss.NewStyle().Foreground(muted).Render(strings.ToUpper(ds.ModeLabel))
	context := ds.Project
	if ds.Task != "" {
		if context != "" {
			context += " · "
		}
		context += ds.Task
	}
	context = TruncateText(context, width-TextWidth(formatClock(ds))-TextWidth(ds.ModeLabel)-8, "…")
	right := lipgloss.NewStyle().Foreground(txt).Render(context)
	line := left + lipgloss.NewStyle().Foreground(muted).Render("  ") + label
	if context != "" {
		line += lipgloss.NewStyle().Foreground(muted).Render("  │  ") + right
	}
	bar := ProgressBar(ds.Progress, width, lipgloss.Color(th.ProgressFill.String()), lipgloss.Color(th.ProgressTrack.String()))
	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, line+"\n"+bar)
}
