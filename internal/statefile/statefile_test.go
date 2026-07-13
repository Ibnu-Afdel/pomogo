package statefile

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func useTempRuntimeDir(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
}

// TestNewManager tests Manager creation.
func TestNewManager(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.statePath == "" {
		t.Error("State path should not be empty")
	}
}

// TestWrite tests writing state to file.
func TestWrite(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Remove()

	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})

	err = manager.Write(session)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Check file exists
	_, err = os.Stat(manager.statePath)
	if err != nil {
		t.Errorf("State file not created: %v", err)
	}
}

// TestRead tests reading state from file.
func TestRead(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Remove()

	// Write initial state
	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})
	if err := manager.Write(session); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read it back
	state, err := manager.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if state == nil {
		t.Fatal("Read returned nil state")
	}
	if state.SessionState != "work" {
		t.Errorf("SessionState = %q, want %q", state.SessionState, "work")
	}
	if state.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", state.PID, os.Getpid())
	}
}

// TestReadNonexistent tests reading when file doesn't exist.
func TestReadNonexistent(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Remove file if it exists
	_ = manager.Remove()

	state, err := manager.Read()
	if err != nil {
		t.Fatalf("Read should not error on missing file: %v", err)
	}
	if state != nil {
		t.Error("Read should return nil for missing file")
	}
}

// TestRemove tests removing state file.
func TestRemove(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	// Write state
	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})
	if err := manager.Write(session); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Remove it
	if err := manager.Remove(); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Check it's gone
	_, err = os.Stat(manager.statePath)
	if !os.IsNotExist(err) {
		t.Error("State file should be removed")
	}
}

// TestStatePathFormat tests state path format.
func TestStatePathFormat(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	path := manager.StatePath()
	if !filepath.IsAbs(path) {
		t.Error("State path should be absolute")
	}
	if filepath.Base(path) != "state.json" {
		t.Errorf("State file should be named state.json, got %s", filepath.Base(path))
	}
}

// TestSessionStateString tests state string conversion.
func TestSessionStateString(t *testing.T) {
	tests := []struct {
		state timer.SessionState
		want  string
	}{
		{timer.StateIdle, "idle"},
		{timer.StateWork, "work"},
		{timer.StateShortBreak, "short_break"},
		{timer.StateLongBreak, "long_break"},
	}

	for _, tt := range tests {
		got := sessionStateString(tt.state)
		if got != tt.want {
			t.Errorf("sessionStateString(%v) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// TestSessionPhaseString tests phase string conversion.
func TestSessionPhaseString(t *testing.T) {
	tests := []struct {
		phase timer.SessionPhase
		want  string
	}{
		{timer.PhaseWork, "work"},
		{timer.PhaseShortBreak, "short_break"},
		{timer.PhaseLongBreak, "long_break"},
	}

	for _, tt := range tests {
		got := sessionPhaseString(tt.phase)
		if got != tt.want {
			t.Errorf("sessionPhaseString(%v) = %q, want %q", tt.phase, got, tt.want)
		}
	}
}

// TestIsExpired tests expiration checking.
func TestIsExpired(t *testing.T) {
	tests := []struct {
		name     string
		state    *State
		expected bool
	}{
		{
			"nil state",
			nil,
			true,
		},
		{
			"future end time",
			&State{EndsAt: time.Now().Add(1 * time.Minute).Unix()},
			false,
		},
		{
			"past end time",
			&State{EndsAt: time.Now().Add(-1 * time.Minute).Unix()},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExpired(tt.state)
			if got != tt.expected {
				t.Errorf("IsExpired = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestIsStale tests staleness checking.
func TestIsStale(t *testing.T) {
	tests := []struct {
		name     string
		state    *State
		expected bool
	}{
		{
			"nil state",
			nil,
			true,
		},
		{
			"any valid state",
			&State{PID: os.Getpid()},
			false,
		},
		{
			"missing process",
			&State{PID: 99999999},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsStale(tt.state)
			if got != tt.expected {
				t.Errorf("IsStale = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestAtomicWrite tests that writes are atomic (temp file → rename).
func TestAtomicWrite(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Remove()

	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})

	// Write should not leave temp files
	if err := manager.Write(session); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	tempPath := manager.statePath + ".tmp"
	_, err = os.Stat(tempPath)
	if !os.IsNotExist(err) {
		t.Error("Temp file should not exist after successful write")
	}
}

// TestRoundtrip tests writing and reading back state.
func TestRoundtrip(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Remove()

	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})
	session.SessionCount = 3

	// Write
	if err := manager.Write(session); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read back
	state, err := manager.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if state.SessionState != "work" {
		t.Errorf("SessionState = %q, want %q", state.SessionState, "work")
	}
	if state.SessionCount != 3 {
		t.Errorf("SessionCount = %d, want 3", state.SessionCount)
	}
	if state.PID != os.Getpid() {
		t.Errorf("PID mismatch")
	}
}

// BenchmarkWrite benchmarks state file writes.
func BenchmarkWrite(b *testing.B) {
	b.Setenv("XDG_RUNTIME_DIR", b.TempDir())
	manager, err := NewManager()
	if err != nil {
		b.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Remove()

	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Write(session)
	}
}

// BenchmarkRead benchmarks state file reads.
func BenchmarkRead(b *testing.B) {
	b.Setenv("XDG_RUNTIME_DIR", b.TempDir())
	manager, err := NewManager()
	if err != nil {
		b.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Remove()

	session := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	session.Start(timer.RealClock{})
	if err := manager.Write(session); err != nil {
		b.Fatalf("Write failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Read()
	}
}
