package ui

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

func TestNewModel(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model == nil {
		t.Fatal("NewModel returned nil")
	}
	if model.cfg != cfg {
		t.Error("Config not set correctly")
	}
	if model.session == nil {
		t.Fatal("Session not initialized")
	}
	if model.theme == nil {
		t.Fatal("Theme not initialized")
	}
}

func TestModelInit(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestViewRender(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked: %v", r)
		}
	}()

	view := model.View()
	if view == "" {
		t.Error("View returned empty string")
	}
}

func TestViewSmallTerminal(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 30
	model.height = 10

	view := model.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("expected 'Terminal too small', got: %q", view)
	}
}

func TestViewTinyTerminal(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 20
	model.height = 5

	view := model.View()
	if !strings.Contains(view, "Terminal too small") {
		t.Errorf("expected 'Terminal too small', got: %q", view)
	}
}

func TestGetDurationForPhase(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	tests := []struct {
		phase    timer.SessionPhase
		expected time.Duration
	}{
		{timer.PhaseWork, cfg.WorkDurationAsDuration()},
		{timer.PhaseShortBreak, cfg.ShortBreakDurationAsDuration()},
		{timer.PhaseLongBreak, cfg.LongBreakDurationAsDuration()},
	}

	for _, tt := range tests {
		model.session.Phase = tt.phase
		got := model.getDurationForPhase()
		if got != tt.expected {
			t.Errorf("getDurationForPhase(%v) = %v, want %v", tt.phase, got, tt.expected)
		}
	}
}

func TestWindowResize(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.width != 80 || model.height != 24 {
		t.Error("Initial dimensions should be 80x24")
	}

	model.width = 120
	model.height = 40

	if model.width != 120 || model.height != 40 {
		t.Error("Dimensions not updated correctly")
	}
}

func TestSessionTracking(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.session.IsRunning {
		t.Error("Session should not be running initially")
	}
}

func TestThemeLoading(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.theme == nil {
		t.Fatal("Theme should be loaded")
	}
	if model.theme.Name == "" {
		t.Error("Theme name should not be empty")
	}
}

func TestMultipleModels(t *testing.T) {
	cfg1 := config.Default()
	cfg2 := config.Default()

	m1 := NewModel(cfg1)
	m2 := NewModel(cfg2)

	if m1 == m2 {
		t.Error("Each NewModel should create a new instance")
	}
	if m1.session == m2.session {
		t.Error("Each model should have its own session")
	}
}

func TestPhaseLabel(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	// Idle state
	label := phaseLabel(model.session)
	if label != "Focus Timer" {
		t.Errorf("idle label = %q, want %q", label, "Focus Timer")
	}

	// Start a work session
	_ = model.session.Start(timer.RealClock{})
	label = phaseLabel(model.session)
	if label != "Work" {
		t.Errorf("work label = %q, want %q", label, "Work")
	}

	// Skip to short break
	model.session.Skip()
	label = phaseLabel(model.session)
	if label != "Short Break" {
		t.Errorf("short break label = %q, want %q", label, "Short Break")
	}
}

func TestBigClockRows(t *testing.T) {
	rows := bigClockRows("00:00", lipgloss.Color("#ff0000"))
	if len(rows) != 5 {
		t.Errorf("bigClockRows returned %d rows, want 5", len(rows))
	}
	for i, row := range rows {
		if row == "" {
			t.Errorf("row %d is empty", i)
		}
	}
}

func TestProgressBar(t *testing.T) {
	bar := progressBar(0.5, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	if bar == "" {
		t.Error("progressBar returned empty string")
	}

	// Test edge cases
	_ = progressBar(0.0, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	_ = progressBar(1.0, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	_ = progressBar(1.5, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	_ = progressBar(-0.5, 10, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
}

func TestSessionDots(t *testing.T) {
	dots := sessionDots(2, 4, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	if dots == "" {
		t.Error("sessionDots returned empty string")
	}

	// Test zero completed
	dots = sessionDots(0, 4, lipgloss.Color("#ff0000"), lipgloss.Color("#333333"))
	if dots == "" {
		t.Error("sessionDots returned empty string for 0 completed")
	}
}

func BenchmarkRender(b *testing.B) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 80
	model.height = 24

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.View()
	}
}

func BenchmarkGetDurationForPhase(b *testing.B) {
	cfg := config.Default()
	model := NewModel(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.getDurationForPhase()
	}
}

func TestRecordSession(t *testing.T) {
	// Set XDG_DATA_HOME to a temp directory so we don't overwrite the real DB
	tempDir, err := os.MkdirTemp("", "pomogo-ui-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldXDGDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tempDir)
	defer func() {
		if oldXDGDataHome == "" {
			os.Unsetenv("XDG_DATA_HOME")
		} else {
			os.Setenv("XDG_DATA_HOME", oldXDGDataHome)
		}
	}()

	cfg := config.Default()
	model := NewModel(cfg)
	if model.dbStore == nil {
		t.Fatal("expected dbStore to be initialized, got nil")
	}

	// Record a completed work session
	now := time.Now().Truncate(time.Second)
	model.recordSession(timer.PhaseWork, now.Add(-25*time.Minute), now, true, 25*time.Minute)

	// Retrieve sessions from dbStore
	sessions, err := model.dbStore.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session to be recorded, got %d", len(sessions))
	}

	got := sessions[0]
	if got.Type != "work" {
		t.Errorf("expected type 'work', got %q", got.Type)
	}
	if !got.Completed {
		t.Errorf("expected completed to be true")
	}
	if got.DurationSecs != 1500 {
		t.Errorf("expected duration_secs to be 1500, got %d", got.DurationSecs)
	}

	// Verify that break sessions are NOT recorded
	model.recordSession(timer.PhaseShortBreak, now, now.Add(5*time.Minute), true, 5*time.Minute)
	sessions2, err := model.dbStore.GetSessions(now.Add(-1*time.Hour), now.Add(1*time.Hour))
	if err != nil {
		t.Fatalf("failed to query sessions: %v", err)
	}
	if len(sessions2) != 1 {
		t.Errorf("expected break session to be ignored, got count %d", len(sessions2))
	}
}
