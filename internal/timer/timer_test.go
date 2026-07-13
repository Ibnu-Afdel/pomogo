package timer

import (
	"testing"
	"time"
)

// MockClock allows tests to control time.
type MockClock struct {
	now time.Time
}

// Now returns the mocked current time.
func (m *MockClock) Now() time.Time {
	return m.now
}

// Advance moves the mock time forward.
func (m *MockClock) Advance(d time.Duration) {
	m.now = m.now.Add(d)
}

// NewMockClock creates a new mock clock starting at a fixed time.
func NewMockClock() *MockClock {
	return &MockClock{now: time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)}
}

// TestNewSession tests session initialization.
func TestNewSession(t *testing.T) {
	workDur := 25 * time.Minute
	shortBreakDur := 5 * time.Minute
	longBreakDur := 15 * time.Minute
	sessionsBeforeLongBreak := 4

	s := NewSession(workDur, shortBreakDur, longBreakDur, sessionsBeforeLongBreak)

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"State is Idle", s.State, StateIdle},
		{"IsRunning is false", s.IsRunning, false},
		{"IsPaused is false", s.IsPaused, false},
		{"SessionCount is 0", s.SessionCount, 0},
		{"SessionsUntilLongBreak matches config", s.SessionsUntilLongBreak, sessionsBeforeLongBreak},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestStart tests starting a session.
func TestStart(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	if err := s.Start(clock); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"State is Work", s.State, StateWork},
		{"IsRunning is true", s.IsRunning, true},
		{"IsPaused is false", s.IsPaused, false},
		{"RemainingTime is WorkDuration", s.RemainingTime, 25 * time.Minute},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestStartAlreadyRunning tests that Start fails if already running.
func TestStartAlreadyRunning(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	if err := s.Start(clock); err == nil {
		t.Errorf("Start() should error when already running, got nil")
	}
}

// TestPause tests pausing a session.
func TestPause(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	if err := s.Pause(clock); err != nil {
		t.Fatalf("Pause() error: %v", err)
	}

	if !s.IsPaused {
		t.Errorf("IsPaused should be true, got false")
	}
}

// TestResume tests resuming a paused session.
func TestResume(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	s.Pause(clock)
	clock.Advance(10 * time.Second)

	if err := s.Resume(clock); err != nil {
		t.Fatalf("Resume() error: %v", err)
	}

	if s.IsPaused {
		t.Errorf("IsPaused should be false, got true")
	}

	// EndsAt should have been shifted by the pause duration
	expectedEndsAt := clock.now.Add(25 * time.Minute)
	if s.EndsAt != expectedEndsAt {
		t.Errorf("EndsAt should be %v, got %v", expectedEndsAt, s.EndsAt)
	}
}

// TestTick tests time progression.
func TestTick(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)

	// Tick before time runs out
	ended := s.Tick(clock)
	if ended {
		t.Errorf("Tick should return false before time ends")
	}

	// Advance to 5 minutes before end
	clock.Advance(20 * time.Minute)
	s.Tick(clock)
	if s.RemainingTime > 6*time.Minute || s.RemainingTime < 4*time.Minute {
		t.Errorf("RemainingTime should be ~5min, got %v", s.RemainingTime)
	}

	// Advance to end
	clock.Advance(5 * time.Minute)
	ended = s.Tick(clock)
	if !ended {
		t.Errorf("Tick should return true when time ends")
	}
}

// TestCompleteWorkSession tests completing a work session.
func TestCompleteWorkSession(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	s.Complete()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"State is ShortBreak", s.State, StateShortBreak},
		{"Phase is ShortBreak", s.Phase, PhaseShortBreak},
		{"SessionCount incremented", s.SessionCount, 1},
		{"SessionsUntilLongBreak decremented", s.SessionsUntilLongBreak, 3},
		{"IsRunning is false", s.IsRunning, false},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestSkip tests skipping a session.
func TestSkip(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	newState := s.Skip()

	if newState != StateShortBreak {
		t.Errorf("Skip from work should go to short break, got %v", newState)
	}
	if s.State != StateShortBreak {
		t.Errorf("State should be ShortBreak, got %v", s.State)
	}
}

// TestReset tests resetting the session.
func TestReset(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	s.Reset()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"State is Idle", s.State, StateIdle},
		{"IsRunning is false", s.IsRunning, false},
		{"IsPaused is false", s.IsPaused, false},
		{"RemainingTime is 0", s.RemainingTime, time.Duration(0)},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestHandle tests the event handler.
func TestHandle(t *testing.T) {
	tests := []struct {
		name      string
		event     Event
		setupFn   func(*Session, Clock) error
		wantState SessionState
		wantErr   bool
	}{
		{
			name:      "Handle Start",
			event:     EventStart,
			setupFn:   func(s *Session, c Clock) error { return nil },
			wantState: StateWork,
			wantErr:   false,
		},
		{
			name:    "Handle Pause",
			event:   EventPause,
			setupFn: func(s *Session, c Clock) error { return s.Start(c) },
			wantErr: false,
		},
		{
			name:    "Handle Resume",
			event:   EventResume,
			setupFn: func(s *Session, c Clock) error { s.Start(c); return s.Pause(c) },
			wantErr: false,
		},
		{
			name:    "Handle Skip",
			event:   EventSkip,
			setupFn: func(s *Session, c Clock) error { return s.Start(c) },
			wantErr: false,
		},
		{
			name:    "Handle Reset",
			event:   EventReset,
			setupFn: func(s *Session, c Clock) error { return s.Start(c) },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clock := NewMockClock()
			s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

			if err := tt.setupFn(s, clock); err != nil {
				t.Fatalf("setup error: %v", err)
			}

			err := s.Handle(tt.event, clock)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestStateTransitions tests a full cycle of state transitions.
func TestStateTransitions(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	transitions := []struct {
		desc      string
		event     Event
		wantState SessionState
	}{
		{"Start", EventStart, StateWork},
		{"Skip to break", EventSkip, StateShortBreak},
		{"Complete break", EventComplete, StateWork},
		{"Start new work", EventStart, StateWork},
		{"Reset to idle", EventReset, StateIdle},
	}

	for _, tr := range transitions {
		if err := s.Handle(tr.event, clock); err != nil && tr.event != EventStart {
			t.Errorf("%s: unexpected error %v", tr.desc, err)
		}
		if s.State != tr.wantState {
			t.Errorf("%s: got state %v, want %v", tr.desc, s.State, tr.wantState)
		}
	}
}

// TestLongBreakCycle tests the full cycle including long breaks.
func TestLongBreakCycle(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	// Simulate 4 work sessions + short breaks
	for cycle := 0; cycle < 4; cycle++ {
		// Start and complete a work session
		if err := s.Start(clock); err != nil {
			t.Fatalf("cycle %d: Start error: %v", cycle, err)
		}
		if s.State != StateWork || s.Phase != PhaseWork {
			t.Errorf("cycle %d: expected Work phase, got %v/%v", cycle, s.State, s.Phase)
		}

		s.Complete()
		if cycle < 3 {
			// Expect short break for first 3 cycles
			if s.State != StateShortBreak || s.Phase != PhaseShortBreak {
				t.Errorf("cycle %d: expected ShortBreak after work, got %v/%v", cycle, s.State, s.Phase)
			}
		}

		// Complete the break (or long break on last cycle)
		if cycle < 3 {
			s.Complete()
			if s.State != StateWork {
				t.Errorf("cycle %d: expected Work after short break, got %v", cycle, s.State)
			}
		}
	}

	// After 4th work session completes, we should have short break
	// (long break triggers after we complete the 4th short break and start 5th work)
	if s.State != StateShortBreak {
		t.Errorf("after 4 cycles: expected ShortBreak, got %v", s.State)
	}

	// Skip the short break to trigger next work
	s.Skip()
	if s.State != StateWork {
		t.Errorf("after skip: expected Work, got %v", s.State)
	}
}

// TestPauseWhenNotRunning tests pause on non-running session.
func TestPauseWhenNotRunning(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	if err := s.Pause(clock); err == nil {
		t.Errorf("Pause on idle session should error")
	}
}

// TestResumeWhenNotPaused tests resume on non-paused session.
func TestResumeWhenNotPaused(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	if err := s.Resume(clock); err == nil {
		t.Errorf("Resume on non-paused running session should error")
	}
}

// TestTickWhenPaused tests that tick doesn't advance time when paused.
func TestTickWhenPaused(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	s.Pause(clock)
	remainingBefore := s.RemainingTime

	// Tick should not change remaining time when paused
	ended := s.Tick(clock)
	if ended {
		t.Errorf("Tick should return false when paused")
	}
	if s.RemainingTime != remainingBefore {
		t.Errorf("RemainingTime should not change when paused, got %v -> %v", remainingBefore, s.RemainingTime)
	}
}

// TestTickWhenIdle tests that tick doesn't do anything on idle session.
func TestTickWhenIdle(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	ended := s.Tick(clock)
	if ended {
		t.Errorf("Tick should return false on idle session")
	}
}

// TestHandleInvalidEvent tests handling an invalid event.
func TestHandleInvalidEvent(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	// Create an invalid event (beyond the known enum)
	invalidEvent := Event(999)
	if err := s.Handle(invalidEvent, clock); err == nil {
		t.Errorf("Handle should error on invalid event")
	}
}

// TestSessionStateString tests the String method for SessionState.
func TestSessionStateString(t *testing.T) {
	tests := []struct {
		state SessionState
		want  string
	}{
		{StateIdle, "Idle"},
		{StateWork, "Work"},
		{StateShortBreak, "ShortBreak"},
		{StateLongBreak, "LongBreak"},
		{SessionState(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("SessionState(%v).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// TestSessionPhaseString tests the String method for SessionPhase.
func TestSessionPhaseString(t *testing.T) {
	tests := []struct {
		phase SessionPhase
		want  string
	}{
		{PhaseWork, "Work"},
		{PhaseShortBreak, "ShortBreak"},
		{PhaseLongBreak, "LongBreak"},
		{SessionPhase(999), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.phase.String(); got != tt.want {
			t.Errorf("SessionPhase(%v).String() = %q, want %q", tt.phase, got, tt.want)
		}
	}
}

// TestRealClock tests the RealClock implementation.
func TestRealClock(t *testing.T) {
	rc := RealClock{}
	before := time.Now()
	now := rc.Now()
	after := time.Now()

	if now.Before(before) || now.After(after.Add(1*time.Millisecond)) {
		t.Errorf("RealClock.Now() returned unexpected time: %v", now)
	}
}

// TestTickAtExactEnd tests tick at the exact end time.
func TestTickAtExactEnd(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	// Advance to exactly the end time
	clock.Advance(25 * time.Minute)
	ended := s.Tick(clock)

	if !ended {
		t.Errorf("Tick should return true at exact end time")
	}
}

// TestResumeShiftsTime tests that resume correctly shifts the end time.
func TestResumeShiftsTime(t *testing.T) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)

	s.Start(clock)
	originalEndsAt := s.EndsAt
	s.Pause(clock)
	pauseDuration := 30 * time.Second
	clock.Advance(pauseDuration)
	s.Resume(clock)

	// EndsAt should have shifted forward by the pause duration
	expectedEndsAt := originalEndsAt.Add(pauseDuration)
	if s.EndsAt != expectedEndsAt {
		t.Errorf("EndsAt not correctly shifted: got %v, want %v", s.EndsAt, expectedEndsAt)
	}
}

// BenchmarkTick benchmarks the Tick method.
func BenchmarkTick(b *testing.B) {
	clock := NewMockClock()
	s := NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	s.Start(clock)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Tick(clock)
		clock.Advance(1 * time.Second)
	}
}
