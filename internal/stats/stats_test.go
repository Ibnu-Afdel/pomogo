package stats

import (
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/store"
)

func TestCalculate(t *testing.T) {
	now := time.Date(2026, 7, 14, 12, 0, 0, 0, time.Local)
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)

	tests := []struct {
		name     string
		sessions []*store.Session
		verify   func(t *testing.T, s *Stats)
	}{
		{
			name:     "empty sessions",
			sessions: []*store.Session{},
			verify: func(t *testing.T, s *Stats) {
				if s.TodayCount != 0 || s.TodayMinutes != 0 || s.WeekCount != 0 || s.MonthCount != 0 {
					t.Errorf("expected 0 for all counts, got %+v", s)
				}
				if s.CurrentStreak != 0 || s.BestStreak != 0 {
					t.Errorf("expected 0 for streaks, got %+v", s)
				}
			},
		},
		{
			name: "today and yesterday sessions",
			sessions: []*store.Session{
				{
					Type:         "work",
					Completed:    true,
					StartedAt:    now.Add(-30 * time.Minute),
					DurationSecs: 1500, // 25 min
				},
				{
					Type:         "work",
					Completed:    true,
					StartedAt:    now.Add(-10 * time.Minute),
					DurationSecs: 1500, // 25 min
				},
				{
					Type:         "work",
					Completed:    false, // incomplete
					StartedAt:    now.Add(-5 * time.Minute),
					DurationSecs: 1500,
				},
				{
					Type:         "work",
					Completed:    true,
					StartedAt:    yesterday,
					DurationSecs: 1500,
				},
				{
					Type:         "short_break", // should be ignored
					Completed:    true,
					StartedAt:    now,
					DurationSecs: 300,
				},
			},
			verify: func(t *testing.T, s *Stats) {
				if s.TodayCount != 2 {
					t.Errorf("expected TodayCount = 2, got %d", s.TodayCount)
				}
				if s.TodayMinutes != 50 { // 25 + 25
					t.Errorf("expected TodayMinutes = 50, got %d", s.TodayMinutes)
				}
				if s.WeekCount != 3 { // 2 today + 1 yesterday
					t.Errorf("expected WeekCount = 3, got %d", s.WeekCount)
				}
				if s.MonthCount != 3 {
					t.Errorf("expected MonthCount = 3, got %d", s.MonthCount)
				}
				// 3 completed work out of 4 total work
				expectedRate := 0.75
				if s.CompletionRate != expectedRate {
					t.Errorf("expected CompletionRate = %v, got %v", expectedRate, s.CompletionRate)
				}
				if s.CurrentStreak != 2 { // today + yesterday
					t.Errorf("expected CurrentStreak = 2, got %d", s.CurrentStreak)
				}
				if s.BestStreak != 2 {
					t.Errorf("expected BestStreak = 2, got %d", s.BestStreak)
				}
			},
		},
		{
			name: "streak calculation with gaps",
			sessions: []*store.Session{
				{
					Type:      "work",
					Completed: true,
					StartedAt: now, // today
				},
				{
					Type:      "work",
					Completed: true,
					StartedAt: yesterday, // yesterday
				},
				// Gap at 2 days ago
				{
					Type:      "work",
					Completed: true,
					StartedAt: now.AddDate(0, 0, -3), // 3 days ago
				},
				{
					Type:      "work",
					Completed: true,
					StartedAt: now.AddDate(0, 0, -4), // 4 days ago
				},
				{
					Type:      "work",
					Completed: true,
					StartedAt: now.AddDate(0, 0, -5), // 5 days ago
				},
			},
			verify: func(t *testing.T, s *Stats) {
				if s.CurrentStreak != 2 { // today + yesterday
					t.Errorf("expected CurrentStreak = 2, got %d", s.CurrentStreak)
				}
				if s.BestStreak != 3 { // 3 days ago through 5 days ago (3 days)
					t.Errorf("expected BestStreak = 3, got %d", s.BestStreak)
				}
			},
		},
		{
			name: "streak alive yesterday (no session today)",
			sessions: []*store.Session{
				{
					Type:      "work",
					Completed: true,
					StartedAt: yesterday, // yesterday
				},
				{
					Type:      "work",
					Completed: true,
					StartedAt: twoDaysAgo, // 2 days ago
				},
			},
			verify: func(t *testing.T, s *Stats) {
				if s.CurrentStreak != 2 { // yesterday + 2 days ago
					t.Errorf("expected CurrentStreak = 2, got %d", s.CurrentStreak)
				}
				if s.BestStreak != 2 {
					t.Errorf("expected BestStreak = 2, got %d", s.BestStreak)
				}
			},
		},
		{
			name: "streak dead (no session today nor yesterday)",
			sessions: []*store.Session{
				{
					Type:      "work",
					Completed: true,
					StartedAt: twoDaysAgo, // 2 days ago
				},
			},
			verify: func(t *testing.T, s *Stats) {
				if s.CurrentStreak != 0 {
					t.Errorf("expected CurrentStreak = 0, got %d", s.CurrentStreak)
				}
				if s.BestStreak != 1 {
					t.Errorf("expected BestStreak = 1, got %d", s.BestStreak)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Calculate(tt.sessions, now)
			tt.verify(t, s)

			// Verify WeekDays order (should be from 6 days ago to today)
			if len(s.WeekDays) != 7 {
				t.Fatalf("expected WeekDays len to be 7, got %d", len(s.WeekDays))
			}
			for i := 0; i < 7; i++ {
				expectedDate := now.AddDate(0, 0, -6+i).Format("2006-01-02")
				gotDate := s.WeekDays[i].Date.Format("2006-01-02")
				if gotDate != expectedDate {
					t.Errorf("WeekDays[%d] date: got %s, want %s", i, gotDate, expectedDate)
				}
			}
		})
	}
}
