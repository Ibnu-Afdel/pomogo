package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestDefaultConfig tests that the default configuration has sensible values.
func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"WorkDuration", cfg.WorkDuration, 25},
		{"ShortBreakDuration", cfg.ShortBreakDuration, 5},
		{"LongBreakDuration", cfg.LongBreakDuration, 15},
		{"SessionsBeforeLongBreak", cfg.SessionsBeforeLongBreak, 4},
		{"Theme", cfg.Theme, "tokyo-night"},
		{"NotificationsEnabled", cfg.NotificationsEnabled, true},
		{"SoundEnabled", cfg.SoundEnabled, true},
		{"PromptForNotes", cfg.PromptForNotes, true},
		{"PauseOnLock", cfg.PauseOnLock, true},
		{"PauseOnSuspend", cfg.PauseOnSuspend, true},
		{"TerminalTitleEnabled", cfg.TerminalTitleEnabled, true},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestDurationConversion tests conversion of config durations to time.Duration.
func TestDurationConversion(t *testing.T) {
	cfg := &Config{
		WorkDuration:       25,
		ShortBreakDuration: 5,
		LongBreakDuration:  15,
	}

	tests := []struct {
		name string
		got  time.Duration
		want time.Duration
	}{
		{"WorkDuration", cfg.WorkDurationAsDuration(), 25 * time.Minute},
		{"ShortBreakDuration", cfg.ShortBreakDurationAsDuration(), 5 * time.Minute},
		{"LongBreakDuration", cfg.LongBreakDurationAsDuration(), 15 * time.Minute},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestValidate tests configuration validation.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "Valid config",
			cfg:     Default(),
			wantErr: false,
		},
		{
			name: "Zero work duration",
			cfg: &Config{
				WorkDuration:            0,
				ShortBreakDuration:      5,
				LongBreakDuration:       15,
				SessionsBeforeLongBreak: 4,
				Theme:                   "tokyo-night",
			},
			wantErr: true,
		},
		{
			name: "Negative short break duration",
			cfg: &Config{
				WorkDuration:            25,
				ShortBreakDuration:      -5,
				LongBreakDuration:       15,
				SessionsBeforeLongBreak: 4,
				Theme:                   "tokyo-night",
			},
			wantErr: true,
		},
		{
			name: "Zero sessions before long break",
			cfg: &Config{
				WorkDuration:            25,
				ShortBreakDuration:      5,
				LongBreakDuration:       15,
				SessionsBeforeLongBreak: 0,
				Theme:                   "tokyo-night",
			},
			wantErr: true,
		},
		{
			name: "Empty theme",
			cfg: &Config{
				WorkDuration:            25,
				ShortBreakDuration:      5,
				LongBreakDuration:       15,
				SessionsBeforeLongBreak: 4,
				Theme:                   "",
			},
			wantErr: true,
		},
		{
			name: "Invalid layout",
			cfg: &Config{
				WorkDuration:            25,
				ShortBreakDuration:      5,
				LongBreakDuration:       15,
				SessionsBeforeLongBreak: 4,
				Theme:                   "tokyo-night",
				Layout:                  "invalid-layout",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestXDGConfigDir tests XDG config directory computation.
func TestXDGConfigDir(t *testing.T) {
	// Save original env
	origHome := os.Getenv("HOME")
	origConfigHome := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("XDG_CONFIG_HOME", origConfigHome)
	}()

	tests := []struct {
		name          string
		xdgConfigHome string
		home          string
		wantSuffix    string
	}{
		{
			name:          "XDG_CONFIG_HOME set",
			xdgConfigHome: "/custom/config",
			home:          "/home/user",
			wantSuffix:    "/custom/config/pomogo",
		},
		{
			name:          "XDG_CONFIG_HOME empty, use HOME",
			xdgConfigHome: "",
			home:          "/home/user",
			wantSuffix:    "/home/user/.config/pomogo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
			os.Setenv("HOME", tt.home)

			got := XDGConfigDir()
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("XDGConfigDir() = %s, want suffix %s", got, tt.wantSuffix)
			}
		})
	}
}

// TestLoadMissingFile tests loading when config file doesn't exist.
func TestLoadMissingFile(t *testing.T) {
	// Create a temporary directory for this test
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Override XDG_CONFIG_HOME to point to our temp directory
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Ensure config directory doesn't exist
	configDir := XDGConfigDir()
	os.RemoveAll(configDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not error when file missing, got: %v", err)
	}

	// Should return defaults
	if cfg.WorkDuration != 25 {
		t.Errorf("Expected default WorkDuration, got %d", cfg.WorkDuration)
	}
}

// TestWriteDefault tests writing default config file.
func TestWriteDefault(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	err := WriteDefault()
	if err != nil {
		t.Fatalf("WriteDefault() error: %v", err)
	}

	// Verify file was created
	path := ConfigFilePath()
	_, err = os.Stat(path)
	if err != nil {
		t.Errorf("Config file not created at %s: %v", path, err)
	}

	// Verify file contains expected content
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	content := string(data)
	if !contains(content, "work_duration = 25") {
		t.Errorf("Config file missing work_duration setting")
	}
	if !contains(content, "theme = \"tokyo-night\"") {
		t.Errorf("Config file missing theme setting")
	}
}

// TestWriteDefaultExists tests WriteDefault when file already exists.
func TestWriteDefaultExists(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create the config directory and file
	configDir := XDGConfigDir()
	os.MkdirAll(configDir, 0755)
	configFile := ConfigFilePath()
	os.WriteFile(configFile, []byte("existing content"), 0644)

	// WriteDefault should error since file exists
	err := WriteDefault()
	if err == nil {
		t.Errorf("WriteDefault() should error when file exists, got nil")
	}
}

// TestLoadValidFile tests loading a valid config file.
func TestLoadValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory and file with valid TOML
	configDir := XDGConfigDir()
	os.MkdirAll(configDir, 0755)
	configFile := ConfigFilePath()

	content := `
work_duration = 30
short_break_duration = 10
long_break_duration = 20
sessions_before_long_break = 3
theme = "catppuccin"
notifications_enabled = false
sound_enabled = true
`
	os.WriteFile(configFile, []byte(content), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"WorkDuration", cfg.WorkDuration, 30},
		{"ShortBreakDuration", cfg.ShortBreakDuration, 10},
		{"LongBreakDuration", cfg.LongBreakDuration, 20},
		{"SessionsBeforeLongBreak", cfg.SessionsBeforeLongBreak, 3},
		{"Theme", cfg.Theme, "catppuccin"},
		{"NotificationsEnabled", cfg.NotificationsEnabled, false},
		{"SoundEnabled", cfg.SoundEnabled, true},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestLoadInvalidFile tests loading an invalid config file.
func TestLoadInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory and file with invalid TOML
	configDir := XDGConfigDir()
	os.MkdirAll(configDir, 0755)
	configFile := ConfigFilePath()

	invalidTOML := `
work_duration = this is not a number
`
	os.WriteFile(configFile, []byte(invalidTOML), 0644)

	_, err := Load()
	if err == nil {
		t.Errorf("Load() should error on invalid TOML, got nil")
	}
}

// contains checks if a string contains a substring.
func contains(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 &&
		(needle == haystack || (len(needle) <= len(haystack) &&
			(haystack == needle || findSubstring(haystack, needle))))
}

// findSubstring is a simple substring search.
func findSubstring(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func TestResolveProfile(t *testing.T) {
	cfg := Default()
	
	// Create profiles
	fifty := 50
	ten := 10
	themeStr := "catppuccin"
	projStr := "backend"

	soundEventStr := "complete"
	cfg.Profiles = map[string]Profile{
		"coding": {
			WorkDuration:       &fifty,
			ShortBreakDuration: &ten,
			Theme:              &themeStr,
			Project:            &projStr,
			SoundEvent:         &soundEventStr,
		},
	}

	// 1. Resolve existing profile
	resolved, project, soundEvent := cfg.ResolveProfile("coding")
	if resolved.WorkDuration != 50 {
		t.Errorf("expected WorkDuration 50, got %d", resolved.WorkDuration)
	}
	if resolved.ShortBreakDuration != 10 {
		t.Errorf("expected ShortBreakDuration 10, got %d", resolved.ShortBreakDuration)
	}
	if resolved.Theme != "catppuccin" {
		t.Errorf("expected Theme 'catppuccin', got %q", resolved.Theme)
	}
	if project != "backend" {
		t.Errorf("expected project 'backend', got %q", project)
	}
	if soundEvent != "complete" {
		t.Errorf("expected soundEvent 'complete', got %q", soundEvent)
	}

	// 2. Fallbacks should stay intact
	if resolved.LongBreakDuration != cfg.LongBreakDuration {
		t.Errorf("expected LongBreakDuration fallback to %d, got %d", cfg.LongBreakDuration, resolved.LongBreakDuration)
	}

	// 3. Resolve non-existent profile
	resolvedNon, projectNon, soundEventNon := cfg.ResolveProfile("non-existent")
	if resolvedNon.WorkDuration != cfg.WorkDuration {
		t.Errorf("expected WorkDuration to be unchanged")
	}
	if projectNon != "" {
		t.Errorf("expected empty project name")
	}
	if soundEventNon != "" {
		t.Errorf("expected empty sound event")
	}
}

func TestModeSpecificConfig(t *testing.T) {
	tmpDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := XDGConfigDir()
	os.MkdirAll(configDir, 0755)
	configFile := ConfigFilePath()

	// 1. Test explicit configuration overriding top level
	content := `
work_duration = 25
short_break_duration = 5
theme = "gruvbox"

[quick_focus]
work_duration = 30
short_break_duration = 6
long_break_duration = 12

[deep_focus]
work_duration = 90
short_break_duration = 10
`
	os.WriteFile(configFile, []byte(content), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.QuickFocusWorkDurationAsDuration() != 30*time.Minute {
		t.Errorf("got QuickFocusWorkDuration %v, want 30m", cfg.QuickFocusWorkDurationAsDuration())
	}
	if cfg.QuickFocusShortBreakDurationAsDuration() != 6*time.Minute {
		t.Errorf("got QuickFocusShortBreakDuration %v, want 6m", cfg.QuickFocusShortBreakDurationAsDuration())
	}
	if cfg.QuickFocusLongBreakDurationAsDuration() != 12*time.Minute {
		t.Errorf("got QuickFocusLongBreakDuration %v, want 12m", cfg.QuickFocusLongBreakDurationAsDuration())
	}
	if cfg.DeepFocusWorkDurationAsDuration() != 90*time.Minute {
		t.Errorf("got DeepFocusWorkDuration %v, want 90m", cfg.DeepFocusWorkDurationAsDuration())
	}
	if cfg.DeepFocusShortBreakDurationAsDuration() != 10*time.Minute {
		t.Errorf("got DeepFocusShortBreakDuration %v, want 10m", cfg.DeepFocusShortBreakDurationAsDuration())
	}

	// 2. Test fallbacks
	contentFallback := `
work_duration = 20
short_break_duration = 4
long_break_duration = 10
sessions_before_long_break = 5
`
	os.WriteFile(configFile, []byte(contentFallback), 0644)

	cfgFallback, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfgFallback.QuickFocusWorkDurationAsDuration() != 20*time.Minute {
		t.Errorf("got QuickFocusWorkDuration %v, want 20m fallback", cfgFallback.QuickFocusWorkDurationAsDuration())
	}
	if cfgFallback.DeepFocusWorkDurationAsDuration() != 20*time.Minute {
		t.Errorf("got DeepFocusWorkDuration %v, want 20m fallback", cfgFallback.DeepFocusWorkDurationAsDuration())
	}
	if cfgFallback.QuickFocusSessionsBeforeLongBreak() != 5 {
		t.Errorf("got SessionsBeforeLongBreak %d, want 5 fallback", cfgFallback.QuickFocusSessionsBeforeLongBreak())
	}
}
