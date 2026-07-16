package render

import (
	"fmt"
	"strings"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	Registry["dashboard"] = LayoutSpec{Layout: Dashboard, MinWidth: 60, MinHeight: 16}
}

// Dashboard renders a denser operational view without turning the timer into a stats app.
func Dashboard(ds DisplayState, th *theme.Theme, f Frame) string {
	color := phaseColor(ds.PhaseKind, th)
	muted := lipgloss.Color(th.Muted.String())
	txt := lipgloss.Color(th.Text.String())
	border := lipgloss.Color(th.Border.String())

	width := f.Width - 10
	if width > 76 {
		width = 76
	}
	if width < 50 {
		width = 50
	}
	leftW := width/2 - 2
	rightW := width - leftW - 3

	timerText := formatClock(ds)
	title := lipgloss.NewStyle().Foreground(color).Bold(true).Render(strings.ToUpper(ds.ModeLabel))
	clock := lipgloss.NewStyle().Foreground(color).Bold(true).Render(timerText)
	bar := ProgressBar(ds.Progress, width, lipgloss.Color(th.ProgressFill.String()), lipgloss.Color(th.ProgressTrack.String()))

	meta := []string{
		metaRow("project", ds.Project, leftW, muted, txt),
		metaRow("task", ds.Task, leftW, muted, txt),
		metaRow("theme", ds.ThemeName, leftW, muted, txt),
	}
	if !ds.Zen {
		meta = append(meta, metaRow("status", ds.StatusMessage, leftW, muted, txt))
	}

	block := []string{
		lipgloss.NewStyle().Foreground(muted).Render("FOCUS"),
		title,
		clock,
		"",
		bar,
	}
	right := lipgloss.NewStyle().Width(rightW).Render(strings.Join(meta, "\n"))
	left := lipgloss.NewStyle().Width(leftW).Render(strings.Join(block, "\n"))

	content := lipgloss.JoinHorizontal(lipgloss.Top, left, lipgloss.NewStyle().Foreground(border).Render(" │ "), right)
	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, box)
}

func metaRow(label, value string, width int, muted, txt lipgloss.Color) string {
	if value == "" {
		value = "none"
	}
	value = TruncateText(value, width-11, "…")
	l := lipgloss.NewStyle().Foreground(muted).Render(PadRight(strings.ToUpper(label), 9))
	v := lipgloss.NewStyle().Foreground(txt).Render(value)
	return l + " " + v
}

func phaseColor(phase timer.SessionPhase, th *theme.Theme) lipgloss.Color {
	switch phase {
	case timer.PhaseWork:
		return lipgloss.Color(th.Work.String())
	case timer.PhaseShortBreak:
		return lipgloss.Color(th.Break.String())
	case timer.PhaseLongBreak:
		return lipgloss.Color(th.LongBreak.String())
	default:
		return lipgloss.Color(th.Idle.String())
	}
}

func formatClock(ds DisplayState) string {
	if ds.BlockRemaining > 0 {
		hours := int(ds.BlockRemaining.Hours())
		mins := int(ds.BlockRemaining.Minutes()) % 60
		secs := int(ds.BlockRemaining.Seconds()) % 60
		return fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	}
	mins := int(ds.SegmentRemaining.Minutes())
	secs := int(ds.SegmentRemaining.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}
