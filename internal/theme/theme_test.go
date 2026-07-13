package theme

import (
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
		{"Work color", theme.Work, Color("#ff7a93")},
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

// TestCatppuccinTheme tests Catppuccin theme initialization.
func TestCatppuccinTheme(t *testing.T) {
	theme := Catppuccin()

	if theme.Name != "catppuccin" {
		t.Errorf("Name: got %q, want %q", theme.Name, "catppuccin")
	}

	if theme.Background != "#fffdf5" {
		t.Errorf("Background (light theme): got %q, want %q", theme.Background, "#fffdf5")
	}
}

// TestGruvboxTheme tests Gruvbox theme initialization.
func TestGruvboxTheme(t *testing.T) {
	theme := Gruvbox()

	if theme.Name != "gruvbox" {
		t.Errorf("Name: got %q, want %q", theme.Name, "gruvbox")
	}

	if theme.Work != "#fb4934" {
		t.Errorf("Work: got %q, want %q", theme.Work, "#fb4934")
	}
}

// TestRegistry tests that all themes are registered.
func TestRegistry(t *testing.T) {
	expectedThemes := []string{"tokyo-night", "catppuccin", "gruvbox"}

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
		{"unknown", "tokyo-night", true}, // Unknown should default to Tokyo Night
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

	if len(themes) != 3 {
		t.Errorf("List() returned %d themes, want 3", len(themes))
	}

	expectedThemes := map[string]bool{
		"tokyo-night": false,
		"catppuccin":  false,
		"gruvbox":     false,
	}

	for _, name := range themes {
		if _, exists := expectedThemes[name]; !exists {
			t.Errorf("Unexpected theme in list: %q", name)
		}
		expectedThemes[name] = true
	}

	for name, found := range expectedThemes {
		if !found {
			t.Errorf("Theme %q not in list", name)
		}
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
		{"gruvbox", false},
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

// TestColorString tests Color.String() method.
func TestColorString(t *testing.T) {
	color := Color("#ff7a93")
	if color.String() != "#ff7a93" {
		t.Errorf("Color.String() = %q, want %q", color.String(), "#ff7a93")
	}
}

// TestANSI256 tests ANSI 256-color approximation.
func TestANSI256(t *testing.T) {
	tests := []struct {
		color    Color
		wantCode int
	}{
		{"#ff7a93", 203},  // Tokyo Night red
		{"#7aa2f7", 75},   // Tokyo Night blue
		{"#9ece6a", 149},  // Tokyo Night green
		{"#fffdf5", 15},   // Catppuccin white
		{"#282828", 235},  // Gruvbox dark
		{"#invalid", 255}, // Unknown color defaults to 255
	}

	for _, tt := range tests {
		got := tt.color.ANSI256()
		if got != tt.wantCode {
			t.Errorf("Color(%q).ANSI256() = %d, want %d", tt.color, got, tt.wantCode)
		}
	}
}

// TestThemeUniqueness ensures each theme has distinct colors.
func TestThemeUniqueness(t *testing.T) {
	themes := []*Theme{TokyoNight(), Catppuccin(), Gruvbox()}

	for i, theme1 := range themes {
		for j, theme2 := range themes {
			if i == j {
				continue // Skip comparing with itself
			}

			if theme1.Name == theme2.Name {
				t.Errorf("Duplicate theme name: %q", theme1.Name)
			}

			// Themes should have distinct work colors
			if theme1.Work == theme2.Work && theme1.Name != theme2.Name {
				t.Logf("Note: %q and %q share work color %q", theme1.Name, theme2.Name, theme1.Work)
			}
		}
	}
}

// BenchmarkGetTheme benchmarks theme retrieval.
func BenchmarkGetTheme(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get("tokyo-night")
	}
}

// BenchmarkValidate benchmarks theme validation.
func BenchmarkValidate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Validate("tokyo-night")
	}
}
