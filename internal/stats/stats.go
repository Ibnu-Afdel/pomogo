// Package stats aggregates focus session metrics.
package stats

import (
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/store"
)

// DayStats represents completed sessions for a single day.
type DayStats struct {
	Date    time.Time
	Count   int
	Minutes int
}

// Stats holds aggregated metrics.
type Stats struct {
	TodayCount       int
	TodayMinutes     int
	WeekCount        int
	WeekMinutes      int
	MonthCount       int
	CurrentStreak    int
	BestStreak       int
	CompletionRate   float64
	WeekDays         [7]DayStats
	LifetimeMinutes  int
	LifetimeSessions int
}

// Calculate computes focus statistics from a list of sessions, optionally filtered by project name.
func Calculate(sessions []*store.Session, now time.Time, projectNameFilter ...string) *Stats {
	stats := &Stats{}
	
	filter := ""
	if len(projectNameFilter) > 0 {
		filter = projectNameFilter[0]
	}

	// Initialize last 7 days of WeekDays dates first
	for i := 0; i < 7; i++ {
		d := now.AddDate(0, 0, -6+i)
		stats.WeekDays[i].Date = d
	}

	if len(sessions) == 0 {
		return stats
	}

	localToday := now.Local().Format("2006-01-02")
	localYesterday := now.Local().AddDate(0, 0, -1).Format("2006-01-02")
	thisYear, thisMonth, _ := now.Local().Date()

	// Map to keep track of completed sessions per day
	completedPerDay := make(map[string]int)
	completedMinsPerDay := make(map[string]int)
	
	var totalWorkSessions int
	var completedWorkSessions int

	for _, s := range sessions {
		if s.Type != "work" {
			continue
		}
		if filter != "" && s.ProjectName != filter {
			continue
		}
		totalWorkSessions++
		if !s.Completed {
			continue
		}
		completedWorkSessions++

		localDate := s.StartedAt.Local().Format("2006-01-02")
		completedPerDay[localDate]++
		mins := s.DurationSecs / 60
		completedMinsPerDay[localDate] += mins

		stats.LifetimeSessions++
		stats.LifetimeMinutes += mins

		// Today stats
		if localDate == localToday {
			stats.TodayCount++
			stats.TodayMinutes += mins
		}

		// Month stats (current calendar month)
		y, m, _ := s.StartedAt.Local().Date()
		if y == thisYear && m == thisMonth {
			stats.MonthCount++
		}
	}

	if totalWorkSessions > 0 {
		stats.CompletionRate = float64(completedWorkSessions) / float64(totalWorkSessions)
	}

	// Calculate last 7 days of WeekDays counts & minutes
	for i := 0; i < 7; i++ {
		dateStr := stats.WeekDays[i].Date.Local().Format("2006-01-02")
		stats.WeekDays[i].Count = completedPerDay[dateStr]
		stats.WeekDays[i].Minutes = completedMinsPerDay[dateStr]
		stats.WeekCount += stats.WeekDays[i].Count
		stats.WeekMinutes += stats.WeekDays[i].Minutes
	}

	// Calculate streaks
	hasToday := completedPerDay[localToday] > 0
	hasYesterday := completedPerDay[localYesterday] > 0

	var streakStart string
	if hasToday {
		streakStart = localToday
	} else if hasYesterday {
		streakStart = localYesterday
	}

	if streakStart != "" {
		current := 0
		checkDate, _ := time.Parse("2006-01-02", streakStart)
		for {
			dateStr := checkDate.Format("2006-01-02")
			if completedPerDay[dateStr] > 0 {
				current++
				checkDate = checkDate.AddDate(0, 0, -1)
			} else {
				break
			}
		}
		stats.CurrentStreak = current
	}

	// Best Streak
	if len(completedPerDay) > 0 {
		var earliest time.Time
		first := true
		for dateStr := range completedPerDay {
			t, err := time.Parse("2006-01-02", dateStr)
			if err == nil {
				if first || t.Before(earliest) {
					earliest = t
					first = false
				}
			}
		}

		best := 0
		currentRun := 0
		limit := now.Local()
		for d := earliest; !d.After(limit); d = d.AddDate(0, 0, 1) {
			dateStr := d.Format("2006-01-02")
			if completedPerDay[dateStr] > 0 {
				currentRun++
				if currentRun > best {
					best = currentRun
				}
			} else {
				currentRun = 0
			}
		}
		stats.BestStreak = best
	}

	return stats
}

// CalculateFocusScore computes a focus rating from 1 to 10 based on session events.
func CalculateFocusScore(pauses, skippedBreaks, abandonedSegments int) int {
	score := 10 - pauses - skippedBreaks - 2*abandonedSegments
	if score < 1 {
		return 1
	}
	if score > 10 {
		return 10
	}
	return score
}
