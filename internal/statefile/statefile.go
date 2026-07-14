// Package statefile provides persistent session state storage via JSON.
package statefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// State represents the serialized state of a Pomodoro session.
type State struct {
	SessionState  string `json:"state"`          // "idle", "work", "short_break", "long_break"
	SessionType   string `json:"session_type"`   // "work", "short_break", "long_break"
	EndsAt        int64  `json:"ends_at"`        // Unix timestamp when current session ends
	Paused        bool   `json:"paused"`         // Whether session is paused
	RemainingSecs int    `json:"remaining_secs"` // Seconds remaining in current session
	PID           int    `json:"pid"`            // Process ID of pomogo app
	SessionCount  int    `json:"session_count"`  // Total sessions completed
	UpdatedAt     int64  `json:"updated_at"`     // Timestamp of last update
	StartedAt     int64  `json:"started_at"`     // Unix timestamp when current session started
	TotalSecs     int    `json:"total_secs"`     // Total seconds in the current phase
	Task          string `json:"task,omitempty"` // Active task description
}

// Manager handles reading and writing session state.
type Manager struct {
	statePath string
}

// NewManager creates a new state file manager.
func NewManager() (*Manager, error) {
	runtimeDir := xdgRuntimeDir()
	pomoDir := filepath.Join(runtimeDir, "pomogo")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(pomoDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create pomogo runtime directory: %w", err)
	}

	return &Manager{
		statePath: filepath.Join(pomoDir, "state.json"),
	}, nil
}

// Write atomically writes the session state to disk.
func (m *Manager) Write(session *timer.Session, task ...string) error {
	taskStr := ""
	if len(task) > 0 {
		taskStr = task[0]
	}
	remaining := session.RemainingTime
	if session.IsRunning && !session.IsPaused {
		remaining = time.Until(session.EndsAt)
		if remaining < 0 {
			remaining = 0
		}
	}

	state := State{
		SessionState:  sessionStateString(session.State),
		SessionType:   sessionPhaseString(session.Phase),
		EndsAt:        session.EndsAt.Unix(),
		Paused:        session.IsPaused,
		RemainingSecs: int(remaining.Seconds()),
		PID:           os.Getpid(),
		SessionCount:  session.SessionCount,
		UpdatedAt:     time.Now().Unix(),
		StartedAt:     session.StartedAt.Unix(),
		TotalSecs:     int(totalForPhase(session).Seconds()),
		Task:          taskStr,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write atomically: write to temp file, then rename
	tempFile := m.statePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write temp state file: %w", err)
	}

	if err := os.Rename(tempFile, m.statePath); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to rename state file: %w", err)
	}

	return nil
}

// Read reads the session state from disk if it exists.
func (m *Manager) Read() (*State, error) {
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist yet
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// Remove deletes the state file.
func (m *Manager) Remove() error {
	if err := os.Remove(m.statePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove state file: %w", err)
		}
	}
	return nil
}

// StatePath returns the full path to the state file.
func (m *Manager) StatePath() string {
	return m.statePath
}

// xdgRuntimeDir returns the XDG_RUNTIME_DIR or a fallback.
func xdgRuntimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	}
	// Fallback to ~/.cache
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp"
	}
	return filepath.Join(home, ".cache")
}

// sessionStateString returns the string representation of a session state.
func sessionStateString(state timer.SessionState) string {
	switch state {
	case timer.StateIdle:
		return "idle"
	case timer.StateWork:
		return "work"
	case timer.StateShortBreak:
		return "short_break"
	case timer.StateLongBreak:
		return "long_break"
	default:
		return "unknown"
	}
}

// sessionPhaseString returns the string representation of a session phase.
func sessionPhaseString(phase timer.SessionPhase) string {
	switch phase {
	case timer.PhaseWork:
		return "work"
	case timer.PhaseShortBreak:
		return "short_break"
	case timer.PhaseLongBreak:
		return "long_break"
	default:
		return "unknown"
	}
}

// IsStale checks if a state is stale (the stored PID is no longer running).
// This is a simple check based on whether the PID can be found in the process table.
func IsStale(state *State) bool {
	if state == nil {
		return true
	}

	if state.PID <= 0 {
		return true
	}

	err := syscall.Kill(state.PID, 0)
	return err == syscall.ESRCH
}

// IsExpired checks if a session state has expired based on EndsAt.
func IsExpired(state *State) bool {
	if state == nil {
		return true
	}
	return time.Now().Unix() > state.EndsAt
}

func totalForPhase(session *timer.Session) time.Duration {
	switch session.Phase {
	case timer.PhaseShortBreak:
		return session.ShortBreakDuration
	case timer.PhaseLongBreak:
		return session.LongBreakDuration
	default:
		return session.WorkDuration
	}
}
