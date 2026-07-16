// Package restore provides session recovery from saved state files.
package restore

import (
	"fmt"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// CanRestore checks if a session can be restored from the state file.
func CanRestore() bool {
	manager, err := statefile.NewManager()
	if err != nil {
		return false
	}

	state, err := manager.Read()
	if err != nil || state == nil {
		return false
	}

	if state.SessionState == "idle" || state.SessionState == "" {
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
		_ = manager.Remove()
		return nil, 0, fmt.Errorf("saved session has expired")
	}

	// Create a new session with the same configuration
	sessionObj := timer.NewSession(
		work,
		shortBreak,
		longBreak,
		sessionsBeforeLongBreak,
	)

	// Restore session state
	if err := restoreSessionState(sessionObj, state); err != nil {
		return nil, 0, fmt.Errorf("failed to restore session state: %w", err)
	}

	// Calculate remaining time
	remainingTime := time.Duration(state.RemainingSecs) * time.Second

	return sessionObj, remainingTime, nil
}

// Durations contains mode-specific timing values used to reconstruct a saved runner.
type Durations struct {
	QuickWork                    time.Duration
	QuickShortBreak              time.Duration
	QuickLongBreak               time.Duration
	QuickSessionsBeforeLongBreak int
	QuickAutoAdvance             bool
	DeepWork                     time.Duration
	DeepShortBreak               time.Duration
}

// RestoreRunnerWithDurations reconstructs the runner and its block plan from state.
func RestoreRunnerWithDurations(d Durations) (*session.Runner, error) {
	manager, err := statefile.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create state manager: %w", err)
	}

	state, err := manager.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	if state == nil {
		return nil, fmt.Errorf("no saved state")
	}

	if statefile.IsExpired(state) {
		_ = manager.Remove()
		return nil, fmt.Errorf("saved session has expired")
	}

	var block *session.Block
	if state.Mode == "deep" {
		plannedTotal := time.Duration(state.PlannedTotalSecs) * time.Second
		block = session.NewDeepBlock(plannedTotal, d.DeepWork, d.DeepShortBreak, true)
		block.Index = state.SegmentIndex
		if block.Index >= 0 && block.Index < len(block.Segments) {
			block.CurrentSegment = block.Segments[block.Index]
		}
	} else {
		block = session.NewQuickBlock(d.QuickWork, d.QuickShortBreak, d.QuickLongBreak, d.QuickSessionsBeforeLongBreak, d.QuickAutoAdvance)
		block.SessionCount = state.SessionCount
		var kind session.SegmentKind
		switch state.SessionType {
		case "work":
			kind = session.SegmentKindWork
		case "short_break":
			kind = session.SegmentKindShortBreak
		case "long_break":
			kind = session.SegmentKindLongBreak
		}
		var dur time.Duration
		switch kind {
		case session.SegmentKindWork:
			dur = d.QuickWork
		case session.SegmentKindShortBreak:
			dur = d.QuickShortBreak
		case session.SegmentKindLongBreak:
			dur = d.QuickLongBreak
		}
		block.CurrentSegment = session.Segment{Kind: kind, Duration: dur}
	}

	runner := session.NewRunner(block)
	runner.Timer.IsRunning = true
	runner.Timer.IsPaused = state.Paused
	runner.Timer.SessionCount = state.SessionCount

	switch state.SessionType {
	case "work":
		runner.Timer.Phase = timer.PhaseWork
		runner.Timer.State = timer.StateWork
	case "short_break":
		runner.Timer.Phase = timer.PhaseShortBreak
		runner.Timer.State = timer.StateShortBreak
	case "long_break":
		runner.Timer.Phase = timer.PhaseLongBreak
		runner.Timer.State = timer.StateLongBreak
	}

	runner.Timer.RemainingTime = time.Duration(state.RemainingSecs) * time.Second
	now := time.Now()
	runner.Timer.EndsAt = now.Add(runner.Timer.RemainingTime)
	runner.Timer.StartedAt = now.Add(-1 * (block.CurrentSegment.Duration - runner.Timer.RemainingTime))
	if runner.Timer.IsPaused {
		runner.Timer.PausedAt = now
	}

	return runner, nil
}

// restoreSessionState updates a session with saved state data.
func restoreSessionState(sessionObj *timer.Session, state *statefile.State) error {
	switch state.SessionType {
	case "work":
		sessionObj.Phase = timer.PhaseWork
		sessionObj.State = timer.StateWork
	case "short_break":
		sessionObj.Phase = timer.PhaseShortBreak
		sessionObj.State = timer.StateShortBreak
	case "long_break":
		sessionObj.Phase = timer.PhaseLongBreak
		sessionObj.State = timer.StateLongBreak
	default:
		return fmt.Errorf("unknown session type: %s", state.SessionType)
	}

	sessionObj.IsRunning = true
	sessionObj.IsPaused = state.Paused
	sessionObj.SessionCount = state.SessionCount
	sessionObj.RemainingTime = time.Duration(state.RemainingSecs) * time.Second
	now := time.Now()
	sessionObj.EndsAt = now.Add(sessionObj.RemainingTime)
	sessionObj.StartedAt = now.Add(-1 * (durationForPhase(sessionObj) - sessionObj.RemainingTime))
	if sessionObj.IsPaused {
		sessionObj.PausedAt = now
	}

	return nil
}

func durationForPhase(sessionObj *timer.Session) time.Duration {
	switch sessionObj.Phase {
	case timer.PhaseShortBreak:
		return sessionObj.ShortBreakDuration
	case timer.PhaseLongBreak:
		return sessionObj.LongBreakDuration
	default:
		return sessionObj.WorkDuration
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
