package session

import (
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

type mockClock struct {
	now time.Time
}

var _ timer.Clock = (*mockClock)(nil)

func (m *mockClock) Now() time.Time {
	return m.now
}

func TestBuildDeepPlan(t *testing.T) {
	tests := []struct {
		name     string
		total    time.Duration
		work     time.Duration
		brk      time.Duration
		expected []Segment
	}{
		{
			name:  "1 Hour Block (25/5)",
			total: 60 * time.Minute,
			work:  25 * time.Minute,
			brk:   5 * time.Minute,
			expected: []Segment{
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 30 * time.Minute},
			},
		},
		{
			name:  "2 Hour Block (25/5)",
			total: 120 * time.Minute,
			work:  25 * time.Minute,
			brk:   5 * time.Minute,
			expected: []Segment{
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 30 * time.Minute},
			},
		},
		{
			name:  "4 Hour Block (25/5/15)",
			total: 240 * time.Minute,
			work:  25 * time.Minute,
			brk:   5 * time.Minute,
			expected: []Segment{
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindLongBreak, Duration: 15 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 20 * time.Minute},
			},
		},
		{
			name:  "45 Min Block (25/5)",
			total: 45 * time.Minute,
			work:  25 * time.Minute,
			brk:   5 * time.Minute,
			expected: []Segment{
				{Kind: SegmentKindWork, Duration: 25 * time.Minute},
				{Kind: SegmentKindShortBreak, Duration: 5 * time.Minute},
				{Kind: SegmentKindWork, Duration: 15 * time.Minute},
			},
		},
		{
			name:  "35 Min Block (25/5) - Folds Break and Trailing Work",
			total: 35 * time.Minute,
			work:  25 * time.Minute,
			brk:   5 * time.Minute,
			expected: []Segment{
				{Kind: SegmentKindWork, Duration: 35 * time.Minute},
			},
		},
		{
			name:  "Short Block (20 min)",
			total: 20 * time.Minute,
			work:  25 * time.Minute,
			brk:   5 * time.Minute,
			expected: []Segment{
				{Kind: SegmentKindWork, Duration: 20 * time.Minute},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildDeepPlan(tt.total, tt.work, tt.brk, 15*time.Minute, 4)
			if len(result) != len(tt.expected) {
				t.Fatalf("got plan length %d, want %d", len(result), len(tt.expected))
			}
			for i, seg := range result {
				if seg.Kind != tt.expected[i].Kind || seg.Duration != tt.expected[i].Duration {
					t.Errorf("segment %d: got %s (%v), want %s (%v)", i, seg.Kind, seg.Duration, tt.expected[i].Kind, tt.expected[i].Duration)
				}
			}
		})
	}
}

func TestBuildDeepPlanProperties(t *testing.T) {
	work := 25 * time.Minute
	brk := 5 * time.Minute

	// Test a wide range of durations from 10 minutes to 5 hours
	for m := 10; m <= 300; m++ {
		total := time.Duration(m) * time.Minute
		plan := BuildDeepPlan(total, work, brk, 15*time.Minute, 4)

		if len(plan) == 0 {
			t.Fatalf("Plan empty for total = %v", total)
		}

		// Property 1: Sum of segments == total
		var sum time.Duration
		for _, seg := range plan {
			sum += seg.Duration
		}
		if sum != total {
			t.Errorf("Total = %v: sum of segments is %v, want %v", total, sum, total)
		}

		// Property 2: Last segment is always Work
		last := plan[len(plan)-1]
		if last.Kind != SegmentKindWork {
			t.Errorf("Total = %v: last segment kind is %s, want Work", total, last.Kind)
		}

		// Property 3: No work segment (except potentially the last one if total < work) is less than 10 mins
		for i, seg := range plan {
			if seg.Kind == SegmentKindWork && seg.Duration < 10*time.Minute && i < len(plan)-1 {
				t.Errorf("Total = %v: non-last work segment %d has duration %v (under 10m limit)", total, i, seg.Duration)
			}
		}
	}
}

func TestRunnerQuickMode(t *testing.T) {
	block := NewQuickBlock(25*time.Minute, 5*time.Minute, 15*time.Minute, 4, false)
	runner := NewRunner(block)
	clock := &mockClock{now: time.Now()}

	if runner.Block.CurrentSegment.Kind != SegmentKindWork {
		t.Fatal("Initial segment should be Work")
	}

	_ = runner.Start(clock)

	// Tick forward 25 mins
	clock.now = clock.now.Add(25 * time.Minute)
	evt, ended := runner.Tick(clock)
	if !ended || evt.Type != EventSegmentEnded {
		t.Fatal("Should trigger SegmentEnded event")
	}

	// Verify next segment configuration (Short Break because count = 1)
	if runner.Block.CurrentSegment.Kind != SegmentKindShortBreak {
		t.Errorf("next segment kind is %s, want Short Break", runner.Block.CurrentSegment.Kind)
	}

	// For quick mode with AutoAdvance = false, the timer should stop running until start/resume is triggered
	if runner.Timer.IsRunning {
		t.Error("Timer should stop running without AutoAdvance")
	}
}

func TestRunnerQuickModeAutoAdvance(t *testing.T) {
	block := NewQuickBlock(25*time.Minute, 5*time.Minute, 15*time.Minute, 4, true)
	runner := NewRunner(block)
	clock := &mockClock{now: time.Now()}

	_ = runner.Start(clock)

	clock.now = clock.now.Add(25 * time.Minute)
	evt, ended := runner.Tick(clock)
	if !ended || evt.Type != EventSegmentEnded {
		t.Fatal("Should trigger SegmentEnded event")
	}
	if runner.Block.CurrentSegment.Kind != SegmentKindShortBreak {
		t.Errorf("next segment kind is %s, want Short Break", runner.Block.CurrentSegment.Kind)
	}
	if !runner.Timer.IsRunning {
		t.Error("Timer should keep running with AutoAdvance")
	}

	clock.now = clock.now.Add(5 * time.Minute)
	evt, ended = runner.Tick(clock)
	if !ended || evt.Type != EventSegmentEnded {
		t.Fatal("Should trigger break SegmentEnded event")
	}
	if runner.Block.CurrentSegment.Kind != SegmentKindWork {
		t.Errorf("next segment kind is %s, want Work", runner.Block.CurrentSegment.Kind)
	}
	if !runner.Timer.IsRunning {
		t.Error("Timer should keep running into the next work segment")
	}
}

func TestRunnerDeepModeAutoAdvance(t *testing.T) {
	block := NewDeepBlock(60*time.Minute, 25*time.Minute, 5*time.Minute, 15*time.Minute, 4, true)
	runner := NewRunner(block)
	clock := &mockClock{now: time.Now()}

	_ = runner.Start(clock)

	// Segment 0: Work (25 min)
	clock.now = clock.now.Add(25 * time.Minute)
	evt, ended := runner.Tick(clock)
	if !ended || evt.Type != EventSegmentEnded {
		t.Fatal("Expected Segment 0 end")
	}

	// Since AutoAdvance is true, the timer should automatically remain running
	if !runner.Timer.IsRunning {
		t.Error("Timer should automatically advance and run")
	}
	if runner.Block.CurrentSegment.Kind != SegmentKindShortBreak {
		t.Errorf("Next segment is %s, want Short Break", runner.Block.CurrentSegment.Kind)
	}

	// Segment 1: Short Break (5 min)
	clock.now = clock.now.Add(5 * time.Minute)
	evt, ended = runner.Tick(clock)
	if !ended || evt.Type != EventSegmentEnded {
		t.Fatal("Expected Segment 1 end")
	}

	// Segment 2: Work (30 min)
	clock.now = clock.now.Add(30 * time.Minute)
	evt, ended = runner.Tick(clock)
	if !ended || evt.Type != EventBlockEnded {
		t.Fatal("Expected BlockEnded event at the end of the Deep Focus block")
	}

	if runner.Timer.IsRunning {
		t.Error("Timer should not be running after block ends")
	}
}
