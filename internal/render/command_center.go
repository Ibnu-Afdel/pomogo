package render

import (
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	Registry["command-center"] = LayoutSpec{Layout: CommandCenter, MinWidth: 70, MinHeight: 18}
}

// CommandCenter renders a split view for users who want task context visible.
func CommandCenter(ds DisplayState, th *theme.Theme, f Frame) string {
	color := phaseColor(ds.PhaseKind, th)
	muted := lipgloss.Color(th.Muted.String())
	txt := lipgloss.Color(th.Text.String())
	border := lipgloss.Color(th.Border.String())
	width := f.Width - 8
	if width > 92 {
		width = 92
	}
	if width < 62 {
		width = 62
	}
	leftW := width / 2
	rightW := width - leftW - 3

	timerBlock := []string{
		lipgloss.NewStyle().Foreground(muted).Render(strings.ToUpper(ds.ModeLabel)),
		lipgloss.NewStyle().Foreground(color).Bold(true).Render(formatClock(ds)),
		ProgressBar(ds.Progress, leftW, lipgloss.Color(th.ProgressFill.String()), lipgloss.Color(th.ProgressTrack.String())),
	}
	left := lipgloss.NewStyle().Width(leftW).Render(strings.Join(timerBlock, "\n"))

	project := ds.Project
	if project == "" {
		project = "No project"
	}
	task := ds.Task
	if task == "" {
		task = "No task"
	}
	status := ds.StatusMessage
	if ds.Zen {
		status = ""
	}
	rightLines := []string{
		lipgloss.NewStyle().Foreground(muted).Render("PROJECT"),
		lipgloss.NewStyle().Foreground(txt).Bold(true).Render(TruncateText(project, rightW, "…")),
		"",
		lipgloss.NewStyle().Foreground(muted).Render("TASK"),
		lipgloss.NewStyle().Foreground(txt).Render(TruncateText(task, rightW, "…")),
		"",
		lipgloss.NewStyle().Foreground(muted).Render(TruncateText(status, rightW, "…")),
	}
	right := lipgloss.NewStyle().Width(rightW).Render(strings.Join(rightLines, "\n"))
	divider := lipgloss.NewStyle().Foreground(border).Render("│")

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", divider, " ", right)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Render(body)
	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, box)
}
