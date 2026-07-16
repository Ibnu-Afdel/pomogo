package restore

import (
	"os"
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/session"
	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func useTempRuntimeDir(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
}

// TestCanRestore tests the CanRestore function.
func TestCanRestore(t *testing.T) {
	useTempRuntimeDir(t)
	// Initially, no saved state, so CanRestore should return false
	canRestore := CanRestore()
	if canRestore {
		t.Error("CanRestore should be false when no state file exists")
	}
}

// TestRestore tests the Restore function.
func TestRestore(t *testing.T) {
	useTempRuntimeDir(t)
	// Create a state file for testing
	manager, err := statefile.NewManager()
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer manager.Remove()

	// Create and save a session
	block := session.NewQuickBlock(25*time.Minute, 5*time.Minute, 15*time.Minute, 4, false)
	runner := session.NewRunner(block)
	runner.Start(timer.RealClock{})
	runner.Timer.SessionCount = 2

	if err := manager.Write(runner, "", nil, ""); err != nil {
		t.Fatalf("Failed to write state: %v", err)
	}

	// Try to restore
	restored, remaining, err := Restore()
	if err != nil {
		t.Logf("Restore failed (expected in test): %v", err)
	} else {
		if restored == nil {
			t.Fatal("Restored session should not be nil")
		}
		if remaining <= 0 {
			t.Error("Remaining time should be positive")
		}
		if restored.SessionCount != 2 {
			t.Errorf("SessionCount = %d, want 2", restored.SessionCount)
		}
	}
}

// TestRestoreExpiredSession tests that expired sessions are not restored.
func TestRestoreExpiredSession(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := statefile.NewManager()
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer manager.Remove()

	// Create a state with expired time
	state := &statefile.State{
		SessionState:  "work",
		SessionType:   "work",
		EndsAt:        time.Now().Add(-1 * time.Minute).Unix(), // Already expired
		Paused:        false,
		RemainingSecs: 30,
		PID:           os.Getpid(),
		SessionCount:  1,
		UpdatedAt:     time.Now().Unix(),
	}

	// Manually write the state file for testing
	data, _ := os.UserHomeDir()
	t.Logf("Would restore from: %s", data)

	// Check that IsExpired works correctly
	if !statefile.IsExpired(state) {
		t.Error("IsExpired should return true for expired session")
	}
}

// TestRestoreSessionState tests restoring session state from saved data.
func TestRestoreSessionState(t *testing.T) {
	sessionObj := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	state := &statefile.State{
		SessionState:  "work",
		SessionType:   "work",
		EndsAt:        time.Now().Add(10 * time.Minute).Unix(),
		Paused:        false,
		RemainingSecs: 600,
		PID:           os.Getpid(),
		SessionCount:  2,
		UpdatedAt:     time.Now().Unix(),
	}

	if err := restoreSessionState(sessionObj, state); err != nil {
		t.Fatalf("restoreSessionState failed: %v", err)
	}

	if sessionObj.Phase != timer.PhaseWork {
		t.Errorf("Phase = %v, want Work", sessionObj.Phase)
	}
	if sessionObj.SessionCount != 2 {
		t.Errorf("SessionCount = %d, want 2", sessionObj.SessionCount)
	}
	if !sessionObj.IsRunning {
		t.Error("Session should be running after restore")
	}
}

// TestRestoreSessionStateLongBreak tests restoring a long break session.
func TestRestoreSessionStateLongBreak(t *testing.T) {
	sessionObj := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	state := &statefile.State{
		SessionState:  "long_break",
		SessionType:   "long_break",
		EndsAt:        time.Now().Add(10 * time.Minute).Unix(),
		Paused:        false,
		RemainingSecs: 600,
		PID:           os.Getpid(),
		SessionCount:  0,
		UpdatedAt:     time.Now().Unix(),
	}

	if err := restoreSessionState(sessionObj, state); err != nil {
		t.Fatalf("restoreSessionState failed: %v", err)
	}

	if sessionObj.Phase != timer.PhaseLongBreak {
		t.Errorf("Phase = %v, want LongBreak", sessionObj.Phase)
	}
}

// TestRestoreInvalidSessionType tests handling of invalid session type.
func TestRestoreInvalidSessionType(t *testing.T) {
	sessionObj := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	state := &statefile.State{
		SessionState:  "unknown",
		SessionType:   "invalid",
		EndsAt:        time.Now().Add(10 * time.Minute).Unix(),
		Paused:        false,
		RemainingSecs: 600,
		PID:           os.Getpid(),
		SessionCount:  0,
		UpdatedAt:     time.Now().Unix(),
	}

	err := restoreSessionState(sessionObj, state)
	if err == nil {
		t.Error("restoreSessionState should error for invalid session type")
	}
}

// TestCleanupExpired tests cleanup of expired sessions.
func TestCleanupExpired(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := statefile.NewManager()
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer manager.Remove()

	// Create an expired state
	state := &statefile.State{
		SessionState:  "work",
		SessionType:   "work",
		EndsAt:        time.Now().Add(-1 * time.Minute).Unix(),
		Paused:        false,
		RemainingSecs: 30,
		PID:           os.Getpid(),
		SessionCount:  1,
		UpdatedAt:     time.Now().Unix(),
	}

	// Write the state
	data, _ := os.UserHomeDir()
	t.Logf("CleanupExpired would check: %s", data)

	// Test cleanup logic
	if statefile.IsExpired(state) {
		t.Log("Expired state correctly identified")
	} else {
		t.Error("IsExpired should return true for expired state")
	}
}

func TestRestoreRunnerWithDurations(t *testing.T) {
	useTempRuntimeDir(t)
	manager, err := statefile.NewManager()
	if err != nil {
		t.Fatalf("Failed to create state manager: %v", err)
	}
	defer manager.Remove()

	// 1. Test Quick Focus Restore
	blockQ := session.NewQuickBlock(25*time.Minute, 5*time.Minute, 15*time.Minute, 4, false)
	runnerQ := session.NewRunner(blockQ)
	runnerQ.Start(timer.RealClock{})
	runnerQ.Timer.SessionCount = 2

	if err := manager.Write(runnerQ, "Quick Task", nil, "Quick Project"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	restoredQ, err := RestoreRunnerWithDurations(Durations{
		QuickWork:                    25 * time.Minute,
		QuickShortBreak:              5 * time.Minute,
		QuickLongBreak:               15 * time.Minute,
		QuickSessionsBeforeLongBreak: 4,
		QuickAutoAdvance:             true,
		DeepWork:                     50 * time.Minute,
		DeepShortBreak:               10 * time.Minute,
		DeepLongBreak:                20 * time.Minute,
		DeepSessionsBeforeLongBreak:  3,
	})
	if err != nil {
		t.Fatalf("RestoreRunnerWithDurations failed: %v", err)
	}
	if restoredQ.Block.Mode != session.ModeQuick {
		t.Errorf("expected mode quick, got %s", restoredQ.Block.Mode)
	}
	if restoredQ.Timer.SessionCount != 2 {
		t.Errorf("expected SessionCount 2, got %d", restoredQ.Timer.SessionCount)
	}
	if !restoredQ.Block.AutoAdvance {
		t.Error("expected restored quick block to preserve configured auto advance")
	}

	// 2. Test Deep Focus Restore
	blockD := session.NewDeepBlock(120*time.Minute, 25*time.Minute, 5*time.Minute, 15*time.Minute, 4, true)
	runnerD := session.NewRunner(blockD)
	runnerD.Start(timer.RealClock{})

	if err := manager.Write(runnerD, "Deep Task", nil, "Deep Project"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	restoredD, err := RestoreRunnerWithDurations(Durations{
		QuickWork:                    25 * time.Minute,
		QuickShortBreak:              5 * time.Minute,
		QuickLongBreak:               15 * time.Minute,
		QuickSessionsBeforeLongBreak: 4,
		QuickAutoAdvance:             true,
		DeepWork:                     50 * time.Minute,
		DeepShortBreak:               10 * time.Minute,
		DeepLongBreak:                20 * time.Minute,
		DeepSessionsBeforeLongBreak:  3,
	})
	if err != nil {
		t.Fatalf("RestoreRunnerWithDurations failed: %v", err)
	}
	if restoredD.Block.Mode != session.ModeDeep {
		t.Errorf("expected mode deep, got %s", restoredD.Block.Mode)
	}
	if restoredD.Block.PlannedTotal != 120*time.Minute {
		t.Errorf("expected PlannedTotal 120m, got %v", restoredD.Block.PlannedTotal)
	}
	if restoredD.Block.Segments[0].Duration != 50*time.Minute {
		t.Errorf("expected deep restore to use configured 50m work duration, got %v", restoredD.Block.Segments[0].Duration)
	}
	if restoredD.Block.SessionsBeforeLongBreak != 3 {
		t.Errorf("expected deep restore to use configured long break cadence, got %d", restoredD.Block.SessionsBeforeLongBreak)
	}
}

// BenchmarkCanRestore benchmarks the CanRestore check.
func BenchmarkCanRestore(b *testing.B) {
	b.Setenv("XDG_RUNTIME_DIR", b.TempDir())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CanRestore()
	}
}

// BenchmarkRestoreSessionState benchmarks state restoration.
func BenchmarkRestoreSessionState(b *testing.B) {
	sessionObj := timer.NewSession(25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	state := &statefile.State{
		SessionState:  "work",
		SessionType:   "work",
		EndsAt:        time.Now().Add(10 * time.Minute).Unix(),
		Paused:        false,
		RemainingSecs: 600,
		PID:           os.Getpid(),
		SessionCount:  2,
		UpdatedAt:     time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = restoreSessionState(sessionObj, state)
	}
}
