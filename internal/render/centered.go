package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func init() {
	Registry["centered"] = LayoutSpec{
		Layout:    Centered,
		MinWidth:  50,
		MinHeight: 14,
	}
}

// Centered renders a huge clock at the absolute center, other elements whisper-quiet.
func Centered(ds DisplayState, th *theme.Theme, f Frame) string {
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
	progFill := lipgloss.Color(th.ProgressFill.String())
	progTrack := lipgloss.Color(th.ProgressTrack.String())
	txtCol := lipgloss.Color(th.Text.String())

	textW := f.Width - 6
	if textW > 70 {
		textW = 70
	}
	if textW < 40 {
		textW = 40
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

	// Context headers (quiet uppercase/small font)
	contextParts := []string{strings.ToUpper(ds.ModeLabel)}
	if ds.Project != "" {
		contextParts = append(contextParts, strings.ToUpper(ds.Project))
	}
	headerText := center(lipgloss.NewStyle().Foreground(muted).Bold(true).Render(strings.Join(contextParts, "  ·  ")))

	taskLine := ""
	if ds.Task != "" {
		taskLine = center(lipgloss.NewStyle().Foreground(txtCol).Render(ds.Task))
	}

	status := center(lipgloss.NewStyle().Foreground(muted).Render(ds.StatusMessage))
	bar := ProgressBar(ds.Progress, textW, progFill, progTrack)

	var lines []string
	lines = append(lines, "")
	if !ds.Zen {
		lines = append(lines, headerText)
	} else if ds.Project != "" {
		lines = append(lines, center(lipgloss.NewStyle().Foreground(muted).Bold(true).Render(strings.ToUpper(ds.Project))))
	}
	lines = append(lines, "")
	for _, row := range clockRows {
		lines = append(lines, center(row))
	}
	lines = append(lines, "")
	if taskLine != "" {
		lines = append(lines, taskLine)
	}
	if !ds.Zen {
		lines = append(lines, status)
	}
	lines = append(lines, "")
	lines = append(lines, bar)
	lines = append(lines, "")

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}
