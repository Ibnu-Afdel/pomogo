// Package timer provides the core Pomodoro session state machine.
// It is pure Go with no I/O dependencies. Time is injected via Clock interface.
package timer

import (
	"fmt"
	"time"
)

// SessionState represents the current state of a focus session.
type SessionState int

const (
	StateIdle SessionState = iota
	StateWork
	StateShortBreak
	StateLongBreak
)

// String returns a human-readable name for the state.
func (s SessionState) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateWork:
		return "Work"
	case StateShortBreak:
		return "ShortBreak"
	case StateLongBreak:
		return "LongBreak"
	default:
		return "Unknown"
	}
}

// Event represents an action taken on the session.
type Event int

const (
	EventStart Event = iota
	EventPause
	EventResume
	EventSkip
	EventComplete
	EventReset
)

// Clock is an interface for time operations.
// Allows tests to inject deterministic time.
type Clock interface {
	Now() time.Time
}

// RealClock uses the system clock.
type RealClock struct{}

// Now returns the current time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// Session represents a Pomodoro focus session.
type Session struct {
	State SessionState
	Phase SessionPhase // Work, ShortBreak, or LongBreak

	// Durations (configured at start)
	WorkDuration            time.Duration
	ShortBreakDuration      time.Duration
	LongBreakDuration       time.Duration
	SessionsBeforeLongBreak int

	// Session tracking
	SessionCount           int // Total sessions completed
	SessionsUntilLongBreak int // Sessions until next long break

	// Timing
	StartedAt     time.Time     // When the current session started
	EndsAt        time.Time     // When the current session will end
	PausedAt      time.Time     // When the session was paused (zero if not paused)
	RemainingTime time.Duration // Remaining time in the current session

	// Flags
	IsRunning bool
	IsPaused  bool
}

// SessionPhase indicates the type of the current session.
type SessionPhase int

const (
	PhaseWork SessionPhase = iota
	PhaseShortBreak
	PhaseLongBreak
)

// String returns a human-readable name for the phase.
func (p SessionPhase) String() string {
	switch p {
	case PhaseWork:
		return "Work"
	case PhaseShortBreak:
		return "ShortBreak"
	case PhaseLongBreak:
		return "LongBreak"
	default:
		return "Unknown"
	}
}

// NewSession creates a new idle session with the given durations.
func NewSession(workDuration, shortBreakDuration, longBreakDuration time.Duration, sessionsBeforeLongBreak int) *Session {
	return &Session{
		State:                   StateIdle,
		WorkDuration:            workDuration,
		ShortBreakDuration:      shortBreakDuration,
		LongBreakDuration:       longBreakDuration,
		SessionsBeforeLongBreak: sessionsBeforeLongBreak,
		SessionCount:            0,
		SessionsUntilLongBreak:  sessionsBeforeLongBreak,
		IsRunning:               false,
		IsPaused:                false,
	}
}

// Start begins the currently queued phase. From idle, it starts a work session.
func (s *Session) Start(clock Clock) error {
	if s.IsRunning {
		return fmt.Errorf("session already running")
	}

	if s.State == StateIdle {
		s.Phase = PhaseWork
		s.State = StateWork
	} else {
		s.State = stateForPhase(s.Phase)
	}
	s.IsRunning = true
	s.IsPaused = false

	now := clock.Now()
	s.StartedAt = now
	s.RemainingTime = s.durationForPhase()
	s.EndsAt = now.Add(s.RemainingTime)

	return nil
}

// Pause pauses the current session.
func (s *Session) Pause(clock Clock) error {
	if !s.IsRunning || s.IsPaused {
		return fmt.Errorf("cannot pause: session not running or already paused")
	}

	s.IsPaused = true
	s.PausedAt = clock.Now()
	s.RemainingTime = s.EndsAt.Sub(s.PausedAt)
	if s.RemainingTime < 0 {
		s.RemainingTime = 0
	}
	return nil
}

// Resume resumes a paused session.
func (s *Session) Resume(clock Clock) error {
	if !s.IsRunning || !s.IsPaused {
		return fmt.Errorf("cannot resume: session not paused")
	}

	s.IsPaused = false
	pausedDuration := clock.Now().Sub(s.PausedAt)
	s.StartedAt = s.StartedAt.Add(pausedDuration)
	s.EndsAt = s.EndsAt.Add(pausedDuration)
	s.PausedAt = time.Time{}

	return nil
}

// Tick updates the session state and returns true if the session has ended.
func (s *Session) Tick(clock Clock) bool {
	if !s.IsRunning || s.IsPaused {
		return false
	}

	now := clock.Now()
	if now.After(s.EndsAt) || now.Equal(s.EndsAt) {
		s.Complete()
		return true
	}

	s.RemainingTime = s.EndsAt.Sub(now)
	if s.RemainingTime < 0 {
		s.RemainingTime = 0
	}

	return false
}

// Skip abandons the current session and advances to the next phase.
// Skipping a work session counts toward the long-break cycle, same as completing one.
func (s *Session) Skip() SessionState {
	if s.Phase == PhaseWork {
		s.SessionCount++
		s.SessionsUntilLongBreak--
	}
	s.IsRunning = false
	s.IsPaused = false
	s.RemainingTime = 0
	s.StartedAt = time.Time{}
	s.EndsAt = time.Time{}
	s.PausedAt = time.Time{}
	next := s.advancePhase()
	if next == StateLongBreak {
		s.SessionsUntilLongBreak = s.SessionsBeforeLongBreak
	}
	return next
}

// Complete marks the session as completed and moves to the next phase.
func (s *Session) Complete() SessionState {
	if s.Phase == PhaseWork {
		s.SessionCount++
		s.SessionsUntilLongBreak--
	}

	s.IsRunning = false
	s.IsPaused = false
	next := s.advancePhase()
	if next == StateLongBreak {
		s.SessionsUntilLongBreak = s.SessionsBeforeLongBreak
	}
	return next
}

// advancePhase determines the next phase in the cycle.
func (s *Session) advancePhase() SessionState {
	switch s.Phase {
	case PhaseWork:
		if s.SessionsUntilLongBreak == 0 {
			s.Phase = PhaseLongBreak
			s.State = StateLongBreak
			return StateLongBreak
		}
		s.Phase = PhaseShortBreak
		s.State = StateShortBreak
		return StateShortBreak

	case PhaseShortBreak, PhaseLongBreak:
		s.Phase = PhaseWork
		s.State = StateWork
		return StateWork

	default:
		return StateIdle
	}
}

// Reset resets the session to idle state.
func (s *Session) Reset() {
	s.State = StateIdle
	s.Phase = PhaseWork
	s.IsRunning = false
	s.IsPaused = false
	s.StartedAt = time.Time{}
	s.EndsAt = time.Time{}
	s.PausedAt = time.Time{}
	s.RemainingTime = 0
}

// AddTime adds a duration to the remaining time of a running session.
func (s *Session) AddTime(d time.Duration) {
	if !s.IsRunning {
		return
	}
	s.RemainingTime += d
	if !s.IsPaused {
		s.EndsAt = s.EndsAt.Add(d)
	}
}

func (s *Session) durationForPhase() time.Duration {
	switch s.Phase {
	case PhaseShortBreak:
		return s.ShortBreakDuration
	case PhaseLongBreak:
		return s.LongBreakDuration
	default:
		return s.WorkDuration
	}
}

func stateForPhase(phase SessionPhase) SessionState {
	switch phase {
	case PhaseShortBreak:
		return StateShortBreak
	case PhaseLongBreak:
		return StateLongBreak
	default:
		return StateWork
	}
}

// Handle processes an event and updates the session state.
func (s *Session) Handle(event Event, clock Clock) error {
	switch event {
	case EventStart:
		return s.Start(clock)
	case EventPause:
		return s.Pause(clock)
	case EventResume:
		return s.Resume(clock)
	case EventSkip:
		s.Skip()
		return nil
	case EventComplete:
		s.Complete()
		return nil
	case EventReset:
		s.Reset()
		return nil
	default:
		return fmt.Errorf("unknown event: %v", event)
	}
}
