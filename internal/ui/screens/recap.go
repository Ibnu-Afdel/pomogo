package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// RecapInfo contains metrics to display on the recap screen.
type RecapInfo struct {
	TotalFocused time.Duration
	Segments     int
	Pauses       int
	Streak       int
	IsDeep       bool
	FocusScore   int
}

// Recap renders a beautiful summary of the completed focus session block.
func Recap(width, height int, th *theme.Theme, info RecapInfo) string {
	accentCol := lipgloss.Color(th.Accent.String())
	mutedCol := lipgloss.Color(th.Muted.String())
	txtCol := lipgloss.Color(th.Text.String())
	borderCol := lipgloss.Color(th.Border.String())
	workCol := lipgloss.Color(th.Work.String())

	textW := width - 12
	if textW > 50 {
		textW = 50
	}
	if textW < 34 {
		textW = 34
	}

	center := func(s string) string {
		return lipgloss.NewStyle().Width(textW).Align(lipgloss.Center).Render(s)
	}

	title := lipgloss.NewStyle().Foreground(accentCol).Bold(true).Render("✦ FOCUS CYCLE COMPLETE ✦")
	banner := lipgloss.NewStyle().Foreground(workCol).Bold(true).Render("EXCELLENT WORK!")

	// Stats rows
	modeStr := "Quick Focus"
	if info.IsDeep {
		modeStr = "Deep Focus"
	}

	hrs := int(info.TotalFocused.Hours())
	mins := int(info.TotalFocused.Minutes()) % 60
	secs := int(info.TotalFocused.Seconds()) % 60
	var timeStr string
	if hrs > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hrs, mins)
	} else if mins > 0 {
		timeStr = fmt.Sprintf("%dm %ds", mins, secs)
	} else {
		timeStr = fmt.Sprintf("%ds", secs)
	}

	row := func(label, value string) string {
		labelStyle := lipgloss.NewStyle().Foreground(mutedCol).Width(20).Align(lipgloss.Right)
		valueStyle := lipgloss.NewStyle().Foreground(txtCol).Bold(true).Width(20).Align(lipgloss.Left)
		return labelStyle.Render(strings.ToUpper(label)) + "  " + valueStyle.Render(value)
	}

	streakStr := fmt.Sprintf("%d Days", info.Streak)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, center(title))
	lines = append(lines, center(banner))
	lines = append(lines, "")
	lines = append(lines, center(row("Mode", modeStr)))
	lines = append(lines, center(row("Time Focused", timeStr)))
	lines = append(lines, center(row("Segments Done", fmt.Sprintf("%d Completed", info.Segments))))
	lines = append(lines, center(row("Pauses Taken", fmt.Sprintf("%d", info.Pauses))))
	lines = append(lines, center(row("Daily Streak", streakStr)))
	if info.IsDeep {
		fillBlocks := info.FocusScore
		if fillBlocks < 0 {
			fillBlocks = 0
		}
		if fillBlocks > 10 {
			fillBlocks = 10
		}
		emptyBlocks := 10 - fillBlocks
		scoreVal := strings.Repeat("■", fillBlocks) + strings.Repeat("□", emptyBlocks) + fmt.Sprintf(" %d/10", info.FocusScore)
		lines = append(lines, center(row("Focus Score", scoreVal)))
	}
	lines = append(lines, "")
	lines = append(lines, center(lipgloss.NewStyle().Foreground(mutedCol).Render("Press [Enter] to dismiss")))
	lines = append(lines, "")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(1, 4).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
