package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func init() {
	Registry["classic"] = LayoutSpec{
		Layout:    Classic,
		MinWidth:  50,
		MinHeight: 16,
	}
}

// Classic renders the main timer display inside a themed border box (original layout).
func Classic(ds DisplayState, th *theme.Theme, f Frame) string {
	var colorHex string
	switch ds.PhaseKind {
	case timer.PhaseWork:
		colorHex = th.Work.String()
	case timer.PhaseShortBreak:
		colorHex = th.Break.String()
	case timer.PhaseLongBreak:
		colorHex = th.LongBreak.String()
	default:
		colorHex = th.Idle.String()
	}
	color := lipgloss.Color(colorHex)
	muted := lipgloss.Color(th.Muted.String())
	borderCol := lipgloss.Color(th.Border.String())
	progFill := lipgloss.Color(th.ProgressFill.String())
	progTrack := lipgloss.Color(th.ProgressTrack.String())
	txtCol := lipgloss.Color(th.Text.String())

	// Text area width: total terminal width minus border (2), padding (2×4=8), outer margin (2)
	textW := f.Width - 12
	if textW > 68 {
		textW = 68
	}
	if textW < 34 {
		textW = 34
	}

	center := func(s string) string {
		return lipgloss.NewStyle().Width(textW).Align(lipgloss.Center).Render(s)
	}

	// Time display
	var clockStr string
	if ds.BlockRemaining > 0 {
		hours := int(ds.BlockRemaining.Hours())
		mins := int(ds.BlockRemaining.Minutes()) % 60
		secs := int(ds.BlockRemaining.Seconds()) % 60
		clockStr = fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
	} else {
		mins := int(ds.SegmentRemaining.Minutes())
		secs := int(ds.SegmentRemaining.Seconds()) % 60
		clockStr = fmt.Sprintf("%02d:%02d", mins, secs)
	}
	clockRows := BigClockRows(clockStr, color)

	// Phase and status
	label := lipgloss.NewStyle().Foreground(color).Bold(true).Render(ds.ModeLabel)
	dots := SessionDots(ds.SegmentIndex, ds.SegmentCount, color, muted)
	status := lipgloss.NewStyle().Foreground(muted).Render(ds.StatusMessage)

	// Progress bar
	bar := ProgressBar(ds.Progress, textW, progFill, progTrack)

	// Hint line
	hints := lipgloss.NewStyle().Foreground(muted).Render(
		"s start  ·  space pause  ·  n skip  ·  t task  ·  p project  ·  r reset  ·  ? help  ·  q quit",
	)

	projectLine := ""
	if ds.Project != "" {
		projectLine = lipgloss.NewStyle().Foreground(txtCol).Bold(true).Render("Project: " + ds.Project)
	}

	taskLine := ""
	if ds.Task != "" {
		taskLine = lipgloss.NewStyle().Foreground(txtCol).Italic(true).Render("Task: " + ds.Task)
	}

	gitLine := ""
	if ds.GitBranch != "" && !ds.Zen {
		gitLine = lipgloss.NewStyle().Foreground(muted).Render(" " + ds.GitBranch)
	}

	tmuxLine := ""
	if ds.TmuxSession != "" && !ds.Zen {
		tmuxLine = lipgloss.NewStyle().Foreground(muted).Render("tmux:" + ds.TmuxSession)
	}

	var lines []string
	lines = append(lines, "")
	for _, row := range clockRows {
		lines = append(lines, center(row))
	}
	lines = append(lines, "")
	lines = append(lines, center(label))
	if ds.SegmentCount > 0 && !ds.Zen {
		lines = append(lines, center(dots))
	}
	if projectLine != "" {
		lines = append(lines, center(projectLine))
	}
	if taskLine != "" {
		lines = append(lines, center(taskLine))
	}
	if gitLine != "" {
		lines = append(lines, center(gitLine))
	}
	if tmuxLine != "" {
		lines = append(lines, center(tmuxLine))
	}
	if !ds.Zen {
		lines = append(lines, center(status))
	}
	lines = append(lines, "")
	lines = append(lines, bar)
	if !ds.Zen {
		lines = append(lines, "")
		lines = append(lines, center(hints))
	}
	lines = append(lines, "")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(0, 4).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, box)
}
