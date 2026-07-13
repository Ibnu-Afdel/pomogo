// Package theme provides color themes for the PomoGo UI.
package theme

import (
	"fmt"
)

// Color represents a hex color value.
type Color string

// Theme represents a complete color scheme for the TUI.
type Theme struct {
	Name        string
	Work        Color // Primary color for work sessions
	Break       Color // Primary color for breaks
	LongBreak   Color // Primary color for long breaks
	Idle        Color // Color for idle state
	Accent      Color // Accent color for highlights
	Background  Color // Background color
	Text        Color // Text/foreground color
	Muted       Color // Muted/secondary text
	Description string
}

// Registry maps theme names to Theme instances.
var Registry = map[string]*Theme{
	"tokyo-night": TokyoNight(),
	"catppuccin":  Catppuccin(),
	"gruvbox":     Gruvbox(),
}

// Get retrieves a theme by name. Returns default (Tokyo Night) if not found.
func Get(name string) *Theme {
	if theme, exists := Registry[name]; exists {
		return theme
	}
	// Default to Tokyo Night
	return TokyoNight()
}

// List returns all available theme names.
func List() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}

// TokyoNight returns the Tokyo Night theme (default).
// Dark, moody, with vibrant accents inspired by Tokyo's nightlife.
func TokyoNight() *Theme {
	return &Theme{
		Name:        "tokyo-night",
		Work:        "#ff7a93", // Coral red
		Break:       "#7aa2f7", // Sky blue
		LongBreak:   "#9ece6a", // Spring green
		Idle:        "#565f89", // Charcoal blue
		Accent:      "#bb9af7", // Purple
		Background:  "#1a1b26", // Deep navy
		Text:        "#c0caf5", // Light blue-gray
		Muted:       "#565f89", // Muted blue-gray
		Description: "Dark theme inspired by Tokyo's neon-lit nights. Default theme.",
	}
}

// Catppuccin returns the Catppuccin (Latte) theme.
// Light, pastel-friendly theme inspired by the Catppuccin color palette.
func Catppuccin() *Theme {
	return &Theme{
		Name:        "catppuccin",
		Work:        "#d20f39", // Flamingo red
		Break:       "#1e66f5", // Sapphire blue
		LongBreak:   "#40a02b", // Green
		Idle:        "#9ca0b0", // Overlay gray
		Accent:      "#8839ef", // Mauve purple
		Background:  "#fffdf5", // Latte cream
		Text:        "#4c4f69", // Dark text
		Muted:       "#9ca0b0", // Muted text
		Description: "Light, pastel theme from the Catppuccin palette. Great for daytime.",
	}
}

// Gruvbox returns the Gruvbox (Dark) theme.
// Warm, earthy theme with good contrast inspired by Gruvbox.
func Gruvbox() *Theme {
	return &Theme{
		Name:        "gruvbox",
		Work:        "#fb4934", // Bright red
		Break:       "#83a598", // Aqua/teal
		LongBreak:   "#b8bb26", // Lime green
		Idle:        "#928374", // Gray
		Accent:      "#d3869b", // Purple/mauve
		Background:  "#282828", // Dark background
		Text:        "#ebdbb2", // Light beige
		Muted:       "#928374", // Muted gray
		Description: "Warm, earthy theme with strong contrast. Good for extended use.",
	}
}

// Validate checks if a theme name is valid.
func Validate(name string) error {
	if _, exists := Registry[name]; !exists {
		return fmt.Errorf("unknown theme: %q (available: tokyo-night, catppuccin, gruvbox)", name)
	}
	return nil
}

// String returns a human-readable hex color string.
func (c Color) String() string {
	return string(c)
}

// ANSI256 returns the closest ANSI 256-color approximation for the hex color.
// This is a simplified mapping; a full implementation would use color distance algorithms.
func (c Color) ANSI256() int {
	// Map common theme colors to ANSI 256 palette
	// For simplicity, using a subset of colors
	hexToANSI := map[string]int{
		// Tokyo Night
		"#ff7a93": 203, // bright red
		"#7aa2f7": 75,  // bright blue
		"#9ece6a": 149, // green
		"#565f89": 60,  // blue-gray
		"#bb9af7": 141, // purple
		"#1a1b26": 235, // very dark gray
		"#c0caf5": 189, // light blue-gray
		// Catppuccin
		"#d20f39": 167, // red
		"#1e66f5": 33,  // blue
		"#40a02b": 35,  // green
		"#9ca0b0": 145, // gray
		"#8839ef": 135, // purple
		"#fffdf5": 15,  // white
		"#4c4f69": 59,  // dark gray
		// Gruvbox
		"#fb4934": 208, // bright red
		"#83a598": 108, // aqua
		"#b8bb26": 106, // lime
		"#928374": 102, // gray
		"#d3869b": 168, // mauve
		"#282828": 235, // very dark gray
		"#ebdbb2": 223, // beige
	}

	if code, exists := hexToANSI[string(c)]; exists {
		return code
	}

	// Default to 255 (white) if color not found
	return 255
}
