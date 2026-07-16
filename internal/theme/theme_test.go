package theme

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTokyoNightTheme tests Tokyo Night theme initialization.
func TestTokyoNightTheme(t *testing.T) {
	theme := TokyoNight()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Name", theme.Name, "tokyo-night"},
		{"Work color", theme.Work, Color("#f7768e")},
		{"Break color", theme.Break, Color("#7aa2f7")},
		{"LongBreak color", theme.LongBreak, Color("#9ece6a")},
		{"Idle color", theme.Idle, Color("#565f89")},
		{"Accent color", theme.Accent, Color("#bb9af7")},
		{"Background color", theme.Background, Color("#1a1b26")},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

// TestRegistry tests that all 10 themes are registered.
func TestRegistry(t *testing.T) {
	expectedThemes := []string{
		"tokyo-night", "catppuccin", "catppuccin-latte", "gruvbox",
		"rose-pine", "everforest", "nord", "dracula", "kanagawa", "carbon",
	}

	for _, name := range expectedThemes {
		if _, exists := Registry[name]; !exists {
			t.Errorf("Theme %q not registered", name)
		}
	}

	if len(Registry) != len(expectedThemes) {
		t.Errorf("Registry has %d themes, want %d", len(Registry), len(expectedThemes))
	}
}

// TestGet retrieves themes by name.
func TestGet(t *testing.T) {
	tests := []struct {
		name       string
		wantName   string
		wantExists bool
	}{
		{"tokyo-night", "tokyo-night", true},
		{"catppuccin", "catppuccin", true},
		{"gruvbox", "gruvbox", true},
		{"rose-pine", "rose-pine", true},
		{"unknown", "tokyo-night", true}, // Unknown defaults to Tokyo Night
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme := Get(tt.name)
			if theme.Name != tt.wantName {
				t.Errorf("Get(%q).Name = %q, want %q", tt.name, theme.Name, tt.wantName)
			}
		})
	}
}

// TestList returns all theme names.
func TestList(t *testing.T) {
	themes := List()

	if len(themes) != 10 {
		t.Errorf("List() returned %d themes, want 10", len(themes))
	}
}

// TestValidate tests theme name validation.
func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"tokyo-night", false},
		{"catppuccin", false},
		{"rose-pine", false},
		{"random", false},
		{"daily", false},
		{"invalid-theme", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

// TestExternalThemes tests loading external TOML themes.
func TestExternalThemes(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	themesDir := filepath.Join(tmpDir, "pomogo", "themes")
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		t.Fatalf("failed to create themes dir: %v", err)
	}

	content := `
name = "custom-theme"
work = "#ff0011"
break = "#00ff22"
long-break = "#0000ff"
idle = "#444444"
accent = "#ff00ff"
background = "#111111"
text = "#eeeeee"
muted = "#777777"
subtle = "#222222"
border = "#333333"
progress-fill = "#ff00ff"
progress-track = "#111111"
ambient = "#0a0a0a"
description = "A neat custom theme."
`
	err := os.WriteFile(filepath.Join(themesDir, "custom.toml"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write test theme file: %v", err)
	}

	if err := LoadExternalThemes(); err != nil {
		t.Fatalf("LoadExternalThemes failed: %v", err)
	}

	theme := Registry["custom-theme"]
	if theme == nil {
		t.Fatal("expected 'custom-theme' to be loaded and registered")
	}

	if theme.Work != "#ff0011" {
		t.Errorf("got Work color %s, want #ff0011", theme.Work)
	}
	if theme.LongBreak != "#0000ff" {
		t.Errorf("got LongBreak color %s, want #0000ff", theme.LongBreak)
	}
}
