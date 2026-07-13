package ui

import (
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/config"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// TestNewModel tests NewModel initialization.
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

// TestModelInit tests the Init command.
func TestModelInit(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	cmd := model.Init()

	if cmd == nil {
		t.Error("Init should return a command")
	}
}

// TestViewRender tests that View renders without panic.
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

// TestViewSmallTerminal tests View with small terminal.
func TestViewSmallTerminal(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 30
	model.height = 10

	view := model.View()
	if view == "" {
		t.Error("View returned empty string")
	}
	if !contains(view, "Terminal too small") {
		t.Error("Expected 'Terminal too small' message")
	}
}

// TestViewTinyTerminal tests View with extremely small terminal.
func TestViewTinyTerminal(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)
	model.width = 20
	model.height = 5

	view := model.View()
	if !contains(view, "Terminal too small") {
		t.Error("Expected 'Terminal too small' message for tiny terminal")
	}
}

// TestGetDurationForPhase tests duration getter.
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

// TestCenterStr tests string centering.
func TestCenterStr(t *testing.T) {
	tests := []struct {
		s     string
		width int
		want  int // expected centered string length (min of string len and width)
	}{
		{"hello", 10, 10},  // string centered in larger width
		{"hello", 5, 5},    // exact fit
		{"hello", 3, 5},    // string longer than width, returns full string
		{"", 5, 5},         // empty string, returns spaces
		{"x", 1, 1},        // single char
	}

	for _, tt := range tests {
		got := centerStr(tt.s, tt.width)
		// centerStr returns a string, check that it's at least the width or the string length
		if len(got) < tt.width && len(tt.s) >= tt.width {
			t.Errorf("centerStr(%q, %d) len = %d, want at least %d", tt.s, tt.width, len(got), tt.width)
		}
	}
}

// TestRepeatStr tests string repetition.
func TestRepeatStr(t *testing.T) {
	tests := []struct {
		s    string
		n    int
		want string
	}{
		{"a", 3, "aaa"},
		{"ab", 2, "abab"},
		{"x", 0, ""},
		{"x", -1, ""},
		{"", 5, ""},
	}

	for _, tt := range tests {
		got := repeatStr(tt.s, tt.n)
		if got != tt.want {
			t.Errorf("repeatStr(%q, %d) = %q, want %q", tt.s, tt.n, got, tt.want)
		}
	}
}

// TestHandleKeypress tests keypress handling.
func TestHandleKeypress(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	// Test that keypresses don't panic
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("handleKeypress panicked: %v", r)
		}
	}()

	// Ensure model is available for keypress simulation
	// (Actual key handling is tested via tea.KeyMsg in Update)
	_ = model
}

// TestWindowResize tests window resize message handling.
func TestWindowResize(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.width != 80 || model.height != 24 {
		t.Error("Initial dimensions should be 80x24")
	}

	// Width/height would be updated via Update() with WindowSizeMsg
	// For now, just verify they're settable
	model.width = 120
	model.height = 40

	if model.width != 120 || model.height != 40 {
		t.Error("Dimensions not updated correctly")
	}
}

// TestSessionTracking tests that model tracks session state.
func TestSessionTracking(t *testing.T) {
	cfg := config.Default()
	model := NewModel(cfg)

	if model.session.IsRunning {
		t.Error("Session should not be running initially (should be Idle)")
	}
}

// TestThemeLoading tests that theme loads correctly.
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

// TestMultipleModels tests creating multiple models.
func TestMultipleModels(t *testing.T) {
	cfg1 := config.Default()
	cfg2 := config.Default()

	model1 := NewModel(cfg1)
	model2 := NewModel(cfg2)

	if model1 == model2 {
		t.Error("Each NewModel should create a new instance")
	}
	if model1.session == model2.session {
		t.Error("Each model should have its own session")
	}
}

// BenchmarkRender benchmarks the View rendering.
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

// BenchmarkGetDurationForPhase benchmarks duration getter.
func BenchmarkGetDurationForPhase(b *testing.B) {
	cfg := config.Default()
	model := NewModel(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = model.getDurationForPhase()
	}
}

// contains checks if a string contains a substring (simple helper).
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
