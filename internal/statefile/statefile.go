// Package statefile provides persistent session state storage via JSON.
package statefile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// State represents the serialized state of a Pomodoro session.
type State struct {
	Version            int    `json:"version"`                // 2 for v2 schema
	Mode               string `json:"mode"`                   // "quick", "deep"
	SessionState       string `json:"state"`                  // "idle", "work", "short_break", "long_break"
	SessionType        string `json:"session_type"`           // "work", "short_break", "long_break"
	EndsAt             int64  `json:"ends_at"`                // Unix timestamp when current session ends
	Paused             bool   `json:"paused"`                 // Whether session is paused
	RemainingSecs      int    `json:"remaining_secs"`         // Seconds remaining in current session
	PID                int    `json:"pid"`                    // Process ID of pomogo app
	SessionCount       int    `json:"session_count"`          // Total sessions completed
	UpdatedAt          int64  `json:"updated_at"`             // Timestamp of last update
	StartedAt          int64  `json:"started_at"`             // Unix timestamp when current session started
	TotalSecs          int    `json:"total_secs"`             // Total seconds in the current phase
	Task               string `json:"task,omitempty"`         // Active task description
	ProjectID          *int64 `json:"project_id,omitempty"`   // Active project ID
	ProjectName        string `json:"project_name,omitempty"` // Active project name
	BlockEndsAt        int64  `json:"block_ends_at"`          // Unix timestamp when deep block ends
	BlockRemainingSecs int    `json:"block_remaining_secs"`   // Total remaining block seconds
	SegmentIndex       int    `json:"segment_index"`          // Current segment index
	SegmentCount       int    `json:"segment_count"`          // Total segment count in deep block
	BlockID            int64  `json:"block_id,omitempty"`     // Active database block ID
	PlannedTotalSecs   int    `json:"planned_total_secs"`     // Planned total block seconds
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
func (m *Manager) Write(runner *session.Runner, task string, projectID *int64, projectName string, blockID ...*int64) error {
	remaining := runner.Timer.RemainingTime
	if runner.Timer.IsRunning && !runner.Timer.IsPaused {
		remaining = time.Until(runner.Timer.EndsAt)
		if remaining < 0 {
			remaining = 0
		}
	}

	var blockEndsAt int64
	var blockRemainingSecs int
	if runner.Block.Mode == session.ModeDeep {
		blockRemaining := runner.Block.Remaining(remaining)
		blockRemainingSecs = int(blockRemaining.Seconds())
		blockEndsAt = time.Now().Add(blockRemaining).Unix()
	}

	var dbBlockID int64
	if len(blockID) > 0 && blockID[0] != nil {
		dbBlockID = *blockID[0]
	}

	state := State{
		Version:            2,
		Mode:               string(runner.Block.Mode),
		SessionState:       sessionStateString(runner.Timer.State),
		SessionType:        sessionPhaseString(runner.Timer.Phase),
		EndsAt:             runner.Timer.EndsAt.Unix(),
		Paused:             runner.Timer.IsPaused,
		RemainingSecs:      int(remaining.Seconds()),
		PID:                os.Getpid(),
		SessionCount:       runner.Timer.SessionCount,
		UpdatedAt:          time.Now().Unix(),
		StartedAt:          runner.Timer.StartedAt.Unix(),
		TotalSecs:          int(runner.Block.CurrentSegment.Duration.Seconds()),
		Task:               task,
		ProjectID:          projectID,
		ProjectName:        projectName,
		BlockEndsAt:        blockEndsAt,
		BlockRemainingSecs: blockRemainingSecs,
		SegmentIndex:       runner.Block.Index,
		SegmentCount:       len(runner.Block.Segments),
		BlockID:            dbBlockID,
		PlannedTotalSecs:   int(runner.Block.PlannedTotal.Seconds()),
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

// IsExpired checks if a session state has expired.
func IsExpired(state *State) bool {
	if state == nil {
		return true
	}
	if state.Mode == "deep" && state.BlockEndsAt > 0 {
		return time.Now().Unix() > state.BlockEndsAt
	}
	return time.Now().Unix() > state.EndsAt
}
