package render

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func init() {
	Registry["compact"] = LayoutSpec{
		Layout:    Compact,
		MinWidth:  40,
		MinHeight: 8,
	}
}

// Compact renders a very short timer widget designed for split-screen layouts.
func Compact(ds DisplayState, th *theme.Theme, f Frame) string {
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

	timerDisp := lipgloss.NewStyle().Foreground(color).Bold(true).Render(clockStr)

	// Combine details onto a single line to keep it compact
	var detailParts []string
	if ds.Project != "" {
		detailParts = append(detailParts, ds.Project)
	}
	if ds.Task != "" {
		detailParts = append(detailParts, ds.Task)
	}
	details := strings.Join(detailParts, " · ")
	detailsDisp := ""
	if details != "" {
		detailsDisp = lipgloss.NewStyle().Foreground(txtCol).Render(details)
	}

	statusDisp := lipgloss.NewStyle().Foreground(muted).Render(ds.StatusMessage)

	// Construct compact one-line status
	statusParts := []string{timerDisp}
	if detailsDisp != "" {
		statusParts = append(statusParts, detailsDisp)
	}
	if !ds.Zen {
		statusParts = append(statusParts, statusDisp)
	}

	bar := ProgressBar(ds.Progress, textW, progFill, progTrack)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, center(strings.Join(statusParts, "  |  ")))
	lines = append(lines, "")
	lines = append(lines, bar)
	lines = append(lines, "")

	return lipgloss.Place(f.Width, f.Height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}
