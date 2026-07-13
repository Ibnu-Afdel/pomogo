// Package restore provides session recovery from saved state files.
package restore

import (
	"fmt"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// CanRestore checks if a session can be restored from the state file.
// A session can be restored if:
// 1. A state file exists
// 2. The session has not expired
// 3. The original process is dead (PID no longer running)
func CanRestore() bool {
	manager, err := statefile.NewManager()
	if err != nil {
		return false
	}

	state, err := manager.Read()
	if err != nil || state == nil {
		return false
	}

	// Check if session is expired
	if statefile.IsExpired(state) {
		return false
	}

	// Check if original PID is dead
	if !statefile.IsStale(state) {
		return false // Original process still running
	}

	return true
}

// Restore attempts to restore a session from the saved state file.
// Returns the restored session and remaining time, or an error if restoration fails.
func Restore() (*timer.Session, time.Duration, error) {
	return RestoreWithDurations(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
}

// RestoreWithDurations attempts to restore a session using the active config durations.
func RestoreWithDurations(work, shortBreak, longBreak time.Duration, sessionsBeforeLongBreak int) (*timer.Session, time.Duration, error) {
	manager, err := statefile.NewManager()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create state manager: %w", err)
	}

	state, err := manager.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read state file: %w", err)
	}

	if state == nil {
		return nil, 0, fmt.Errorf("no saved state")
	}

	// Check if session is expired
	if statefile.IsExpired(state) {
		// Clean up expired state
		_ = manager.Remove()
		return nil, 0, fmt.Errorf("saved session has expired")
	}

	// Create a new session with the same configuration
	session := timer.NewSession(
		work,
		shortBreak,
		longBreak,
		sessionsBeforeLongBreak,
	)

	// Restore session state
	if err := restoreSessionState(session, state); err != nil {
		return nil, 0, fmt.Errorf("failed to restore session state: %w", err)
	}

	// Calculate remaining time
	remainingTime := time.Duration(state.RemainingSecs) * time.Second

	return session, remainingTime, nil
}

// restoreSessionState updates a session with saved state data.
func restoreSessionState(session *timer.Session, state *statefile.State) error {
	// Restore phase
	switch state.SessionType {
	case "work":
		session.Phase = timer.PhaseWork
		session.State = timer.StateWork
	case "short_break":
		session.Phase = timer.PhaseShortBreak
		session.State = timer.StateShortBreak
	case "long_break":
		session.Phase = timer.PhaseLongBreak
		session.State = timer.StateLongBreak
	default:
		return fmt.Errorf("unknown session type: %s", state.SessionType)
	}

	// Restore running state
	session.IsRunning = true
	session.IsPaused = state.Paused

	// Restore session count
	session.SessionCount = state.SessionCount

	// Restore timing (we'll set EndsAt based on remaining time)
	// Since we don't know the exact clock offset, we estimate based on remaining time
	session.RemainingTime = time.Duration(state.RemainingSecs) * time.Second
	now := time.Now()
	session.EndsAt = now.Add(session.RemainingTime)
	session.StartedAt = now.Add(-1 * (durationForPhase(session) - session.RemainingTime))
	if session.IsPaused {
		session.PausedAt = now
	}

	return nil
}

func durationForPhase(session *timer.Session) time.Duration {
	switch session.Phase {
	case timer.PhaseShortBreak:
		return session.ShortBreakDuration
	case timer.PhaseLongBreak:
		return session.LongBreakDuration
	default:
		return session.WorkDuration
	}
}

// CleanupExpired removes expired state files.
func CleanupExpired() error {
	manager, err := statefile.NewManager()
	if err != nil {
		return err
	}

	state, err := manager.Read()
	if err != nil {
		return err
	}

	if state != nil && statefile.IsExpired(state) {
		return manager.Remove()
	}

	return nil
}
