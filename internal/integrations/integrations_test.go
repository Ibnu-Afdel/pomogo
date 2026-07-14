package integrations

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/statefile"
)

func TestFormatStatus_Idle(t *testing.T) {
	// Test nil state (idle)
	res, err := FormatStatus(nil, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "Idle" {
		t.Errorf("expected Idle, got %q", res)
	}

	res, err = FormatStatus(nil, "tmux")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "" {
		t.Errorf("expected empty string, got %q", res)
	}

	res, err = FormatStatus(nil, "waybar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var waybar WaybarOutput
	if err := json.Unmarshal([]byte(res), &waybar); err != nil {
		t.Fatalf("failed to parse waybar output: %v", err)
	}
	if waybar.Text != "" || waybar.Class != "pomogo-idle" || waybar.Tooltip != "Idle" {
		t.Errorf("unexpected waybar output: %+v", waybar)
	}

	res, err = FormatStatus(nil, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != `{"state":"idle"}` {
		t.Errorf("expected json idle, got %q", res)
	}
}

func TestFormatStatus_Active(t *testing.T) {
	// Setup an active work state that ends in 15 minutes
	now := time.Now()
	state := &statefile.State{
		SessionState:  "work",
		SessionType:   "work",
		EndsAt:        now.Add(15 * time.Minute).Unix(),
		Paused:        false,
		RemainingSecs: 900,
		PID:           os.Getpid(), // make it not stale
		SessionCount:  2,
		UpdatedAt:     now.Unix(),
		StartedAt:     now.Unix(),
		TotalSecs:     1500,
		Task:          "Code Integrations",
	}

	// Test default format
	res, err := FormatStatus(state, "default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(res, "🍅 15:00 · work") || !strings.Contains(res, "[Code Integrations]") {
		t.Errorf("unexpected default output: %q", res)
	}

	// Test tmux format
	res, err = FormatStatus(state, "tmux")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != "🍅 15:00" {
		t.Errorf("expected '🍅 15:00', got %q", res)
	}

	// Test json format
	res, err = FormatStatus(state, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsedState statefile.State
	if err := json.Unmarshal([]byte(res), &parsedState); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}
	if parsedState.RemainingSecs != 900 || parsedState.Task != "Code Integrations" {
		t.Errorf("unexpected json state fields: %+v", parsedState)
	}

	// Test waybar format
	res, err = FormatStatus(state, "waybar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var waybar WaybarOutput
	if err := json.Unmarshal([]byte(res), &waybar); err != nil {
		t.Fatalf("failed to unmarshal waybar: %v", err)
	}
	if waybar.Text != "🍅 15:00" {
		t.Errorf("expected text '🍅 15:00', got %q", waybar.Text)
	}
	if waybar.Class != "pomogo-work" {
		t.Errorf("expected class 'pomogo-work', got %q", waybar.Class)
	}
	if !strings.Contains(waybar.Tooltip, "Task: Code Integrations") || !strings.Contains(waybar.Tooltip, "Work session ends at") {
		t.Errorf("unexpected tooltip: %q", waybar.Tooltip)
	}
}

func TestFormatStatus_Paused(t *testing.T) {
	now := time.Now()
	state := &statefile.State{
		SessionState:  "work",
		SessionType:   "work",
		EndsAt:        now.Add(15 * time.Minute).Unix(),
		Paused:        true,
		RemainingSecs: 900,
		PID:           os.Getpid(),
		SessionCount:  2,
		UpdatedAt:     now.Unix(),
		StartedAt:     now.Unix(),
		TotalSecs:     1500,
	}

	res, err := FormatStatus(state, "waybar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var waybar WaybarOutput
	if err := json.Unmarshal([]byte(res), &waybar); err != nil {
		t.Fatalf("failed to unmarshal waybar: %v", err)
	}
	if waybar.Text != "⏸️ 15:00" {
		t.Errorf("expected text '⏸️ 15:00', got %q", waybar.Text)
	}
	if waybar.Class != "pomogo-paused" {
		t.Errorf("expected class 'pomogo-paused', got %q", waybar.Class)
	}
}
