package notify

import (
	"testing"

	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// TestNewNotifier tests Notifier creation.
func TestNewNotifier(t *testing.T) {
	notifier := NewNotifier(true, false)
	if notifier == nil {
		t.Fatal("NewNotifier returned nil")
	}
	if !notifier.enabled {
		t.Error("Notifier should be enabled")
	}

	notifierDisabled := NewNotifier(false, false)
	if notifierDisabled.enabled {
		t.Error("Notifier should be disabled")
	}
}

// TestNotifyTransitionDisabled tests that disabled notifier returns early.
func TestNotifyTransitionDisabled(t *testing.T) {
	notifier := NewNotifier(false, false)
	err := notifier.NotifyTransition(timer.StateWork, timer.PhaseWork)
	if err != nil {
		t.Errorf("Disabled notifier should not error: %v", err)
	}
}

// TestNotifyTransitionWork tests work state notification.
func TestNotifyTransitionWork(t *testing.T) {
	notifier := NewNotifier(true, false)
	title, msg, urgency := notifier.messageForTransition(timer.StateWork, timer.PhaseWork)

	if title == "" {
		t.Error("Title should not be empty")
	}
	if msg == "" {
		t.Error("Message should not be empty")
	}
	if urgency != "normal" {
		t.Errorf("Work urgency should be 'normal', got %q", urgency)
	}
}

// TestNotifyTransitionShortBreak tests short break notification.
func TestNotifyTransitionShortBreak(t *testing.T) {
	notifier := NewNotifier(true, false)
	title, msg, urgency := notifier.messageForTransition(timer.StateShortBreak, timer.PhaseShortBreak)

	if title == "" {
		t.Error("Title should not be empty")
	}
	if msg == "" {
		t.Error("Message should not be empty")
	}
	if urgency != "normal" {
		t.Errorf("Short break urgency should be 'normal', got %q", urgency)
	}
}

// TestNotifyTransitionLongBreak tests long break notification.
func TestNotifyTransitionLongBreak(t *testing.T) {
	notifier := NewNotifier(true, false)
	title, msg, urgency := notifier.messageForTransition(timer.StateLongBreak, timer.PhaseLongBreak)

	if title == "" {
		t.Error("Title should not be empty")
	}
	if msg == "" {
		t.Error("Message should not be empty")
	}
	if urgency != "normal" {
		t.Errorf("Long break urgency should be 'normal', got %q", urgency)
	}
}

// TestNotifyTransitionIdle tests idle state notification.
func TestNotifyTransitionIdle(t *testing.T) {
	notifier := NewNotifier(true, false)
	title, msg, urgency := notifier.messageForTransition(timer.StateIdle, timer.PhaseWork)

	if title == "" {
		t.Error("Title should not be empty")
	}
	if msg == "" {
		t.Error("Message should not be empty")
	}
	if urgency != "low" {
		t.Errorf("Idle urgency should be 'low', got %q", urgency)
	}
}

// TestNotifyError tests error notification.
func TestNotifyError(t *testing.T) {
	notifier := NewNotifier(false, false)
	err := notifier.NotifyError("Test error")
	if err != nil {
		t.Errorf("NotifyError should not error even when disabled: %v", err)
	}
}

// TestValidUrgency tests urgency validation.
func TestValidUrgency(t *testing.T) {
	tests := []struct {
		urgency string
		valid   bool
	}{
		{"low", true},
		{"normal", true},
		{"critical", true},
		{"invalid", false},
		{"", false},
		{"HIGH", false}, // case-sensitive
	}

	for _, tt := range tests {
		got := validUrgency(tt.urgency)
		if got != tt.valid {
			t.Errorf("validUrgency(%q) = %v, want %v", tt.urgency, got, tt.valid)
		}
	}
}

// TestStateString tests state string conversion.
func TestStateString(t *testing.T) {
	tests := []struct {
		state timer.SessionState
		want  string
	}{
		{timer.StateWork, "Work"},
		{timer.StateShortBreak, "Short Break"},
		{timer.StateLongBreak, "Long Break"},
		{timer.StateIdle, "Idle"},
	}

	for _, tt := range tests {
		got := StateString(tt.state)
		if got != tt.want {
			t.Errorf("StateString(%v) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// TestNotifyCustom tests custom notification.
func TestNotifyCustom(t *testing.T) {
	notifier := NewNotifier(false, false)
	err := notifier.NotifyCustom("Test", "Message", "normal")
	if err != nil {
		t.Errorf("NotifyCustom should not error: %v", err)
	}
}

// TestNotifyCustomInvalidUrgency tests custom notification with invalid urgency.
func TestNotifyCustomInvalidUrgency(t *testing.T) {
	notifier := NewNotifier(false, false)
	err := notifier.NotifyCustom("Test", "Message", "invalid")
	if err != nil {
		t.Errorf("NotifyCustom should handle invalid urgency: %v", err)
	}
}

// TestNotifyWithMissingNotifySend tests that missing notify-send doesn't crash.
func TestNotifyWithMissingNotifySend(t *testing.T) {
	// This test checks that sending a notification doesn't crash
	// when notify-send is missing (which is the case in test environments).
	notifier := NewNotifier(true, false)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Notifier panicked with missing notify-send: %v", r)
		}
	}()

	// Should not crash even if notify-send is not available
	err := notifier.NotifyTransition(timer.StateWork, timer.PhaseWork)
	if err != nil {
		t.Logf("NotifyTransition returned error (may be expected if notify-send missing): %v", err)
	}
}

// BenchmarkNotifyTransition benchmarks notification sending.
func BenchmarkNotifyTransition(b *testing.B) {
	notifier := NewNotifier(false, false) // Disabled to avoid sending actual notifications

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = notifier.NotifyTransition(timer.StateWork, timer.PhaseWork)
	}
}

// BenchmarkMessageForTransition benchmarks message generation.
func BenchmarkMessageForTransition(b *testing.B) {
	notifier := NewNotifier(true, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = notifier.messageForTransition(timer.StateWork, timer.PhaseWork)
	}
}
