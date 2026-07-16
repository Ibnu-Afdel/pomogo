package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func init() {
	Registry["minimal"] = LayoutSpec{
		Layout:    Minimal,
		MinWidth:  40,
		MinHeight: 10,
	}
}

// Minimal renders a simple borderless timer widget.
func Minimal(ds DisplayState, th *theme.Theme, f Frame) string {
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

	textW := f.Width - 4
	if textW > 60 {
		textW = 60
	}
	if textW < 30 {
		textW = 30
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

	// Quiet minimal timer display
	timerDisp := lipgloss.NewStyle().Foreground(color).Bold(true).Render(clockStr)
	labelDisp := lipgloss.NewStyle().Foreground(muted).Render(" · " + ds.ModeLabel)

	header := center(timerDisp + labelDisp)

	details := ""
	if ds.Project != "" && ds.Task != "" {
		details = center(lipgloss.NewStyle().Foreground(txtCol).Render(ds.Project + " → " + ds.Task))
	} else if ds.Project != "" {
		details = center(lipgloss.NewStyle().Foreground(txtCol).Render(ds.Project))
	} else if ds.Task != "" {
		details = center(lipgloss.NewStyle().Foreground(txtCol).Render(ds.Task))
	}

	bar := ProgressBar(ds.Progress, textW, progFill, progTrack)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, header)
	if details != "" {
		lines = append(lines, details)
	}
	if !ds.Zen {
		status := center(lipgloss.NewStyle().Foreground(muted).Render(ds.StatusMessage))
		lines = append(lines, status)
	}
	lines = append(lines, "")
	lines = append(lines, bar)
	lines = append(lines, "")

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}
