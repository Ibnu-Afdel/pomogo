package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func init() {
	Registry["retro"] = LayoutSpec{
		Layout:    Retro,
		MinWidth:  50,
		MinHeight: 16,
	}
}

// Retro renders the timer inside a double-line border box using textured character sets.
func Retro(ds DisplayState, th *theme.Theme, f Frame) string {
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

	label := lipgloss.NewStyle().Foreground(color).Bold(true).Render(strings.ToUpper(ds.ModeLabel))
	dots := SessionDots(ds.SegmentIndex, ds.SegmentCount, color, muted)
	status := lipgloss.NewStyle().Foreground(muted).Render(strings.ToUpper(ds.StatusMessage))

	bar := retroProgressBar(ds.Progress, textW, progFill, progTrack)

	hints := lipgloss.NewStyle().Foreground(muted).Render(
		"S START  ·  SPACE PAUSE  ·  N SKIP  ·  T TASK  ·  P PROJECT  ·  R RESET  ·  ? HELP  ·  Q QUIT",
	)

	projectLine := ""
	if ds.Project != "" {
		projectLine = lipgloss.NewStyle().Foreground(txtCol).Bold(true).Render("PROJECT: " + strings.ToUpper(ds.Project))
	}

	taskLine := ""
	if ds.Task != "" {
		taskLine = lipgloss.NewStyle().Foreground(txtCol).Italic(true).Render("TASK: " + strings.ToUpper(ds.Task))
	}

	gitLine := ""
	if ds.GitBranch != "" && !ds.Zen {
		gitLine = lipgloss.NewStyle().Foreground(muted).Render(" " + strings.ToUpper(ds.GitBranch))
	}

	tmuxLine := ""
	if ds.TmuxSession != "" && !ds.Zen {
		tmuxLine = lipgloss.NewStyle().Foreground(muted).Render("TMUX:" + strings.ToUpper(ds.TmuxSession))
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
		Border(lipgloss.DoubleBorder()).
		BorderForeground(borderCol).
		Padding(0, 4).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, box)
}

func retroProgressBar(progress float64, width int, fill, track lipgloss.TerminalColor) string {
	if width <= 0 {
		return ""
	}
	filledLen := int(progress * float64(width))
	if filledLen < 0 {
		filledLen = 0
	}
	if filledLen > width {
		filledLen = width
	}
	emptyLen := width - filledLen

	filledStr := strings.Repeat("▓", filledLen)
	emptyStr := strings.Repeat("▒", emptyLen)

	filledStyle := lipgloss.NewStyle().Foreground(fill).Render(filledStr)
	emptyStyle := lipgloss.NewStyle().Foreground(track).Render(emptyStr)

	return filledStyle + emptyStyle
}
