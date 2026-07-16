package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/stats"
	"github.com/Ibnu-Afdel/pomogo/internal/store"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

// Stats renders the focus statistics screen.
func Stats(width, height int, th *theme.Theme, s *stats.Stats, statusMessage string, recent []*store.Session) string {
	accentCol := lipgloss.Color(th.Accent.String())
	mutedCol := lipgloss.Color(th.Muted.String())
	borderCol := lipgloss.Color(th.Border.String())
	progFillCol := lipgloss.Color(th.ProgressFill.String())
	progTrackCol := lipgloss.Color(th.ProgressTrack.String())
	txtCol := lipgloss.Color(th.Text.String())

	textW := width - 12
	if textW > 68 {
		textW = 68
	}
	if textW < 34 {
		textW = 34
	}

	center := func(s string) string {
		return lipgloss.NewStyle().Width(textW).Align(lipgloss.Center).Render(s)
	}

	// Today stats
	todayStr := fmt.Sprintf("Today: %d sessions (%d mins focused)", s.TodayCount, s.TodayMinutes)

	// Streak stats
	streakStr := fmt.Sprintf("Streak: %d days (Best: %d days)", s.CurrentStreak, s.BestStreak)

	// Month/Lifetime stats
	monthStr := fmt.Sprintf("This Month: %d sessions (Rate: %.0f%%)", s.MonthCount, s.CompletionRate*100)
	
	lifeHrs := s.LifetimeMinutes / 60
	lifeMins := s.LifetimeMinutes % 60
	lifeStr := fmt.Sprintf("Lifetime: %d sessions (%dh %dm focused)", s.LifetimeSessions, lifeHrs, lifeMins)

	// Week graph title
	graphTitle := lipgloss.NewStyle().Foreground(accentCol).Bold(true).Render("Weekly Focus Activity (mins)")
	graph := weekBarGraph(s.WeekDays, progFillCol, progTrackCol, mutedCol)

	recentTitle := lipgloss.NewStyle().Foreground(accentCol).Bold(true).Render("Recent Work Sessions")

	var recentLines []string
	if len(recent) == 0 {
		recentLines = append(recentLines, lipgloss.NewStyle().Foreground(mutedCol).Render("No sessions recorded yet"))
	} else {
		for _, r := range recent {
			timeStr := r.StartedAt.Local().Format("15:04")
			taskStr := r.Task
			if taskStr == "" {
				taskStr = "[no task]"
			}

			status := "completed"
			statusColor := progFillCol
			if !r.Completed {
				status = "skipped"
				statusColor = mutedCol
			}

			line := fmt.Sprintf("%s  %s (%s)",
				lipgloss.NewStyle().Foreground(mutedCol).Render(timeStr),
				lipgloss.NewStyle().Foreground(txtCol).Bold(true).Render(taskStr),
				lipgloss.NewStyle().Foreground(statusColor).Render(status),
			)
			if r.Note != "" {
				line += lipgloss.NewStyle().Foreground(mutedCol).Render(" - " + r.Note)
			}
			recentLines = append(recentLines, line)
		}
	}

	// Hints/Status
	var hints string
	if statusMessage != "" {
		hints = lipgloss.NewStyle().Foreground(accentCol).Render(statusMessage)
	} else {
		hints = lipgloss.NewStyle().Foreground(mutedCol).Render("Tab timer  ·  y yank stats  ·  ? help  ·  q quit")
	}

	var lines []string
	lines = append(lines, "")
	lines = append(lines, center(lipgloss.NewStyle().Foreground(accentCol).Bold(true).Render("Focus Statistics")))
	lines = append(lines, "")
	lines = append(lines, center(todayStr))
	lines = append(lines, center(streakStr))
	lines = append(lines, center(monthStr))
	lines = append(lines, center(lifeStr))
	lines = append(lines, "")
	lines = append(lines, center(graphTitle))
	lines = append(lines, "")

	// Center the bar graph block
	graphLines := strings.Split(graph, "\n")
	maxLineLen := 0
	for _, l := range graphLines {
		length := lipgloss.Width(l)
		if length > maxLineLen {
			maxLineLen = length
		}
	}
	pad := (textW - maxLineLen) / 2
	if pad < 0 {
		pad = 0
	}
	var paddedGraphLines []string
	for _, l := range graphLines {
		paddedGraphLines = append(paddedGraphLines, strings.Repeat(" ", pad)+l)
	}
	lines = append(lines, strings.Join(paddedGraphLines, "\n"))

	lines = append(lines, "")
	lines = append(lines, center(recentTitle))
	lines = append(lines, "")

	// Center recent sessions block
	maxRecentLen := 0
	for _, l := range recentLines {
		length := lipgloss.Width(l)
		if length > maxRecentLen {
			maxRecentLen = length
		}
	}
	rPad := (textW - maxRecentLen) / 2
	if rPad < 0 {
		rPad = 0
	}
	var paddedRecentLines []string
	for _, l := range recentLines {
		paddedRecentLines = append(paddedRecentLines, strings.Repeat(" ", rPad)+l)
	}
	lines = append(lines, strings.Join(paddedRecentLines, "\n"))

	lines = append(lines, "")
	lines = append(lines, center(hints))
	lines = append(lines, "")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderCol).
		Padding(1, 4).
		Render(strings.Join(lines, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

func weekBarGraph(weekDays [7]stats.DayStats, fill, track, muted lipgloss.Color) string {
	var sb strings.Builder
	for i, wd := range weekDays {
		if i > 0 {
			sb.WriteByte('\n')
		}

		weekdayName := wd.Date.Format("Mon")
		sb.WriteString(lipgloss.NewStyle().Foreground(muted).Render(weekdayName + "  "))

		// 1 block per 10 minutes focused, capped at 15
		barLen := wd.Minutes / 10
		if barLen > 15 {
			barLen = 15
		}

		if barLen > 0 {
			bar := strings.Repeat("█", barLen)
			sb.WriteString(lipgloss.NewStyle().Foreground(fill).Render(bar))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(track).Render("░"))
		}

		sb.WriteString(fmt.Sprintf(" %dm", wd.Minutes))
	}
	return sb.String()
}
