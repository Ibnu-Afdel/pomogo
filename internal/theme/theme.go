// Package theme provides color themes for the PomoGo UI.
package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// Color represents a hex color value.
type Color string

// Theme represents a complete color scheme for the TUI.
type Theme struct {
	Name          string
	Work          Color // Primary color for work sessions
	Break         Color // Primary color for breaks
	LongBreak     Color // Primary color for long breaks
	Idle          Color // Color for idle state
	Accent        Color // Accent color for highlights
	Background    Color // Background color
	Text          Color // Text/foreground color
	Muted         Color // Muted/secondary text
	Subtle        Color // Subtle background/box colors
	Border        Color // Border colors
	ProgressFill  Color // Filled progress bar color
	ProgressTrack Color // Progress bar track color
	Ambient       Color // Very muted color for backgrounds/effects
	Description   string
}

// Registry maps theme names to Theme instances.
var Registry = map[string]*Theme{
	"tokyo-night":      TokyoNight(),
	"catppuccin":       CatppuccinMocha(),
	"catppuccin-latte": CatppuccinLatte(),
	"gruvbox":          Gruvbox(),
	"rose-pine":        RosePine(),
	"everforest":       Everforest(),
	"nord":             Nord(),
	"dracula":          Dracula(),
	"kanagawa":         Kanagawa(),
	"carbon":           Carbon(),
	"night-owl":        NightOwl(),
	"one-dark":         OneDark(),
	"ayu-mirage":       AyuMirage(),
	"solarized-dark":   SolarizedDark(),
	"oxocarbon":        Oxocarbon(),
	"high-contrast":    HighContrast(),
	"github-dark":      GitHubDark(),
	"material-ocean":   MaterialOcean(),
	"forest-dawn":      ForestDawn(),
}

// Get retrieves a theme by name. Returns default (Tokyo Night) if not found.
func Get(name string) *Theme {
	if theme, exists := Registry[name]; exists {
		return theme
	}
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

// Validate checks if a theme name is valid.
func Validate(name string) error {
	if name == "random" || name == "daily" {
		return nil
	}
	if _, exists := Registry[name]; !exists {
		return fmt.Errorf("unknown theme: %q", name)
	}
	return nil
}

// String returns a human-readable hex color string.
func (c Color) String() string {
	return string(c)
}

// TokyoNight returns the Tokyo Night theme (default).
func TokyoNight() *Theme {
	return &Theme{
		Name:          "tokyo-night",
		Work:          "#f7768e",
		Break:         "#7aa2f7",
		LongBreak:     "#9ece6a",
		Idle:          "#565f89",
		Accent:        "#bb9af7",
		Background:    "#1a1b26",
		Text:          "#c0caf5",
		Muted:         "#565f89",
		Subtle:        "#414868",
		Border:        "#3b4261",
		ProgressFill:  "#bb9af7",
		ProgressTrack: "#24283b",
		Ambient:       "#1f2335",
		Description:   "Dark theme inspired by Tokyo's neon-lit nights.",
	}
}

// CatppuccinMocha returns the Catppuccin Mocha theme.
func CatppuccinMocha() *Theme {
	return &Theme{
		Name:          "catppuccin",
		Work:          "#f38ba8",
		Break:         "#89b4fa",
		LongBreak:     "#a6e3a1",
		Idle:          "#585b70",
		Accent:        "#cba6f7",
		Background:    "#1e1e2e",
		Text:          "#cdd6f4",
		Muted:         "#7f849c",
		Subtle:        "#313244",
		Border:        "#45475a",
		ProgressFill:  "#f38ba8",
		ProgressTrack: "#181825",
		Ambient:       "#11111b",
		Description:   "Dark, warm pastel theme from the Catppuccin palette.",
	}
}

// CatppuccinLatte returns the Catppuccin Latte theme.
func CatppuccinLatte() *Theme {
	return &Theme{
		Name:          "catppuccin-latte",
		Work:          "#d20f39",
		Break:         "#1e66f5",
		LongBreak:     "#40a02b",
		Idle:          "#9ca0b0",
		Accent:        "#8839ef",
		Background:    "#eff1f5",
		Text:          "#4c4f69",
		Muted:         "#7c7f93",
		Subtle:        "#ccd0da",
		Border:        "#acb0be",
		ProgressFill:  "#d20f39",
		ProgressTrack: "#e6e9ef",
		Ambient:       "#ccd0da",
		Description:   "Light pastel theme from the Catppuccin palette.",
	}
}

// Gruvbox returns the Gruvbox (Dark) theme.
func Gruvbox() *Theme {
	return &Theme{
		Name:          "gruvbox",
		Work:          "#fb4934",
		Break:         "#83a598",
		LongBreak:     "#b8bb26",
		Idle:          "#928374",
		Accent:        "#fe8019",
		Background:    "#282828",
		Text:          "#ebdbb2",
		Muted:         "#a89984",
		Subtle:        "#3c3836",
		Border:        "#504945",
		ProgressFill:  "#fb4934",
		ProgressTrack: "#1d2021",
		Ambient:       "#1d2021",
		Description:   "Warm, earthy retro theme with strong contrast.",
	}
}

// RosePine returns the Rose Pine theme.
func RosePine() *Theme {
	return &Theme{
		Name:          "rose-pine",
		Work:          "#ebbcba",
		Break:         "#31748f",
		LongBreak:     "#9ccfd8",
		Idle:          "#6e6a86",
		Accent:        "#c4a7e7",
		Background:    "#191724",
		Text:          "#e0def4",
		Muted:         "#908caa",
		Subtle:        "#26233a",
		Border:        "#403d52",
		ProgressFill:  "#ebbcba",
		ProgressTrack: "#212030",
		Ambient:       "#2a2837",
		Description:   "All natural pine, foam, and rose colors.",
	}
}

// Everforest returns the Everforest theme.
func Everforest() *Theme {
	return &Theme{
		Name:          "everforest",
		Work:          "#e67e80",
		Break:         "#7fbbb3",
		LongBreak:     "#a7c080",
		Idle:          "#859289",
		Accent:        "#dbbc7f",
		Background:    "#2d353b",
		Text:          "#d3c6aa",
		Muted:         "#9da9a0",
		Subtle:        "#343f44",
		Border:        "#475258",
		ProgressFill:  "#a7c080",
		ProgressTrack: "#232a2e",
		Ambient:       "#232a2e",
		Description:   "Warm, comforting forest green palette.",
	}
}

// Nord returns the Nord theme.
func Nord() *Theme {
	return &Theme{
		Name:          "nord",
		Work:          "#bf616a",
		Break:         "#88c0d0",
		LongBreak:     "#a3be8c",
		Idle:          "#4c566a",
		Accent:        "#81a1c1",
		Background:    "#2e3440",
		Text:          "#d8dee9",
		Muted:         "#81a1c1",
		Subtle:        "#3b4252",
		Border:        "#434c5e",
		ProgressFill:  "#88c0d0",
		ProgressTrack: "#2e3440",
		Ambient:       "#3b4252",
		Description:   "Arctic, ice-cold color scheme.",
	}
}

// Dracula returns the Dracula theme.
func Dracula() *Theme {
	return &Theme{
		Name:          "dracula",
		Work:          "#ff5555",
		Break:         "#8be9fd",
		LongBreak:     "#50fa7b",
		Idle:          "#6272a4",
		Accent:        "#bd93f9",
		Background:    "#282a36",
		Text:          "#f8f8f2",
		Muted:         "#6272a4",
		Subtle:        "#343746",
		Border:        "#44475a",
		ProgressFill:  "#bd93f9",
		ProgressTrack: "#191a21",
		Ambient:       "#191a21",
		Description:   "Classic, high-contrast dark theme for vampires.",
	}
}

// Kanagawa returns the Kanagawa theme.
func Kanagawa() *Theme {
	return &Theme{
		Name:          "kanagawa",
		Work:          "#c3404b",
		Break:         "#7e9cd8",
		LongBreak:     "#76946a",
		Idle:          "#727169",
		Accent:        "#957fb8",
		Background:    "#1f1f28",
		Text:          "#dcd7ba",
		Muted:         "#727169",
		Subtle:        "#2a2a37",
		Border:        "#363646",
		ProgressFill:  "#957fb8",
		ProgressTrack: "#16161d",
		Ambient:       "#16161d",
		Description:   "Edo-era artistic theme inspired by the Great Wave.",
	}
}

// Carbon returns the Carbon theme.
func Carbon() *Theme {
	return &Theme{
		Name:          "carbon",
		Work:          "#fa4d56",
		Break:         "#4589ff",
		LongBreak:     "#24a148",
		Idle:          "#525252",
		Accent:        "#a861ea",
		Background:    "#161616",
		Text:          "#f4f4f4",
		Muted:         "#707070",
		Subtle:        "#262626",
		Border:        "#393939",
		ProgressFill:  "#f4f4f4",
		ProgressTrack: "#161616",
		Ambient:       "#262626",
		Description:   "Monochrome, high-contrast gray scheme.",
	}
}

// NightOwl returns the Night Owl theme.
// Palette source: https://github.com/sdras/night-owl-vscode-theme
func NightOwl() *Theme {
	return &Theme{
		Name:          "night-owl",
		Work:          "#ef5350",
		Break:         "#82aaff",
		LongBreak:     "#22da6e",
		Idle:          "#637777",
		Accent:        "#c792ea",
		Background:    "#011627",
		Text:          "#d6deeb",
		Muted:         "#637777",
		Subtle:        "#0b2942",
		Border:        "#1d3b53",
		ProgressFill:  "#82aaff",
		ProgressTrack: "#01111d",
		Ambient:       "#0b253a",
		Description:   "High-contrast blue night palette for late coding.",
	}
}

// OneDark returns the One Dark theme.
// Palette source: https://github.com/atom/atom/tree/master/packages/one-dark-syntax
func OneDark() *Theme {
	return &Theme{
		Name:          "one-dark",
		Work:          "#e06c75",
		Break:         "#61afef",
		LongBreak:     "#98c379",
		Idle:          "#5c6370",
		Accent:        "#c678dd",
		Background:    "#282c34",
		Text:          "#abb2bf",
		Muted:         "#5c6370",
		Subtle:        "#2c313c",
		Border:        "#3e4451",
		ProgressFill:  "#61afef",
		ProgressTrack: "#21252b",
		Ambient:       "#21252b",
		Description:   "Atom's classic balanced dark editor palette.",
	}
}

// AyuMirage returns the Ayu Mirage theme.
// Palette source: https://github.com/ayu-theme/ayu-colors
func AyuMirage() *Theme {
	return &Theme{
		Name:          "ayu-mirage",
		Work:          "#f07178",
		Break:         "#59c2ff",
		LongBreak:     "#bbe67e",
		Idle:          "#607080",
		Accent:        "#ffcc66",
		Background:    "#1f2430",
		Text:          "#cbccc6",
		Muted:         "#707a8c",
		Subtle:        "#242936",
		Border:        "#343f4c",
		ProgressFill:  "#ffcc66",
		ProgressTrack: "#1b1f29",
		Ambient:       "#191e2a",
		Description:   "Warm, low-glare Ayu palette for all-day focus.",
	}
}

// SolarizedDark returns the Solarized Dark theme.
// Palette source: https://ethanschoonover.com/solarized/
func SolarizedDark() *Theme {
	return &Theme{
		Name:          "solarized-dark",
		Work:          "#dc322f",
		Break:         "#268bd2",
		LongBreak:     "#859900",
		Idle:          "#586e75",
		Accent:        "#b58900",
		Background:    "#002b36",
		Text:          "#839496",
		Muted:         "#586e75",
		Subtle:        "#073642",
		Border:        "#657b83",
		ProgressFill:  "#2aa198",
		ProgressTrack: "#073642",
		Ambient:       "#073642",
		Description:   "Precision low-contrast palette built for terminals.",
	}
}

// Oxocarbon returns the Oxocarbon theme.
// Palette source: https://github.com/nyoom-engineering/oxocarbon
func Oxocarbon() *Theme {
	return &Theme{
		Name:          "oxocarbon",
		Work:          "#ee5396",
		Break:         "#33b1ff",
		LongBreak:     "#42be65",
		Idle:          "#525252",
		Accent:        "#be95ff",
		Background:    "#161616",
		Text:          "#f2f4f8",
		Muted:         "#78a9ff",
		Subtle:        "#262626",
		Border:        "#393939",
		ProgressFill:  "#08bdba",
		ProgressTrack: "#262626",
		Ambient:       "#0f0f0f",
		Description:   "IBM-inspired cyberpunk palette with crisp accents.",
	}
}

// HighContrast returns a strict accessibility-oriented dark theme.
func HighContrast() *Theme {
	return &Theme{
		Name:          "high-contrast",
		Work:          "#ff5f5f",
		Break:         "#00d7ff",
		LongBreak:     "#5fff87",
		Idle:          "#bcbcbc",
		Accent:        "#ffd700",
		Background:    "#000000",
		Text:          "#ffffff",
		Muted:         "#bcbcbc",
		Subtle:        "#1c1c1c",
		Border:        "#ffffff",
		ProgressFill:  "#ffd700",
		ProgressTrack: "#3a3a3a",
		Ambient:       "#262626",
		Description:   "Maximum contrast dark theme for readability.",
	}
}

func GitHubDark() *Theme {
	return &Theme{
		Name:          "github-dark",
		Work:          "#ff7b72",
		Break:         "#79c0ff",
		LongBreak:     "#a5d6ff",
		Idle:          "#8b949e",
		Accent:        "#d2a8ff",
		Background:    "#0d1117",
		Text:          "#e6edf3",
		Muted:         "#8b949e",
		Subtle:        "#161b22",
		Border:        "#30363d",
		ProgressFill:  "#58a6ff",
		ProgressTrack: "#30363d",
		Ambient:       "#21262d",
		Description:   "GitHub-inspired dark theme with crisp code-review contrast.",
	}
}

func MaterialOcean() *Theme {
	return &Theme{
		Name:          "material-ocean",
		Work:          "#ff5370",
		Break:         "#82aaff",
		LongBreak:     "#c3e88d",
		Idle:          "#b2ccd6",
		Accent:        "#ffcb6b",
		Background:    "#0f111a",
		Text:          "#eeffff",
		Muted:         "#b2ccd6",
		Subtle:        "#1f2233",
		Border:        "#3b4252",
		ProgressFill:  "#89ddff",
		ProgressTrack: "#2f334d",
		Ambient:       "#252a3a",
		Description:   "Material-style ocean palette with bright but readable accents.",
	}
}

func ForestDawn() *Theme {
	return &Theme{
		Name:          "forest-dawn",
		Work:          "#f26d6d",
		Break:         "#6fb3d2",
		LongBreak:     "#9fd356",
		Idle:          "#c6d3b8",
		Accent:        "#f7c65f",
		Background:    "#10140f",
		Text:          "#edf4e4",
		Muted:         "#c6d3b8",
		Subtle:        "#1d2519",
		Border:        "#46543f",
		ProgressFill:  "#9fd356",
		ProgressTrack: "#2b3327",
		Ambient:       "#283224",
		Description:   "Earthy dark theme with balanced green, blue, and gold signals.",
	}
}

// ExternalTheme represents the TOML representation of an external theme file.
type ExternalTheme struct {
	Name          string `toml:"name"`
	Work          string `toml:"work"`
	Break         string `toml:"break"`
	LongBreak     string `toml:"long-break"`
	Idle          string `toml:"idle"`
	Accent        string `toml:"accent"`
	Background    string `toml:"background"`
	Text          string `toml:"text"`
	Muted         string `toml:"muted"`
	Subtle        string `toml:"subtle"`
	Border        string `toml:"border"`
	ProgressFill  string `toml:"progress-fill"`
	ProgressTrack string `toml:"progress-track"`
	Ambient       string `toml:"ambient"`
	Description   string `toml:"description"`
}

// ToTheme converts ExternalTheme to Theme.
func (et *ExternalTheme) ToTheme() *Theme {
	return &Theme{
		Name:          et.Name,
		Work:          Color(et.Work),
		Break:         Color(et.Break),
		LongBreak:     Color(et.LongBreak),
		Idle:          Color(et.Idle),
		Accent:        Color(et.Accent),
		Background:    Color(et.Background),
		Text:          Color(et.Text),
		Muted:         Color(et.Muted),
		Subtle:        Color(et.Subtle),
		Border:        Color(et.Border),
		ProgressFill:  Color(et.ProgressFill),
		ProgressTrack: Color(et.ProgressTrack),
		Ambient:       Color(et.Ambient),
		Description:   et.Description,
	}
}

// LoadExternalThemes scans for external themes and registers them.
func LoadExternalThemes() error {
	themesDir := filepath.Join(xdgConfigDir(), "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return nil
	}

	files, err := filepath.Glob(filepath.Join(themesDir, "*.toml"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var et ExternalTheme
		if err := toml.Unmarshal(data, &et); err != nil {
			continue
		}
		if et.Name == "" {
			stem := filepath.Base(file)
			stem = strings.TrimSuffix(stem, filepath.Ext(stem))
			et.Name = stem
		}
		Registry[et.Name] = et.ToTheme()
	}
	return nil
}

// CheckExternalThemes returns a list of filenames for malformed themes.
func CheckExternalThemes() []string {
	var malformed []string
	themesDir := filepath.Join(xdgConfigDir(), "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return nil
	}

	files, err := filepath.Glob(filepath.Join(themesDir, "*.toml"))
	if err != nil {
		return nil
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var et ExternalTheme
		if err := toml.Unmarshal(data, &et); err != nil {
			malformed = append(malformed, filepath.Base(file))
		}
	}
	return malformed
}

// CheckExternalThemeContrast returns external theme files with low-contrast core roles.
func CheckExternalThemeContrast() []string {
	var lowContrast []string
	themesDir := filepath.Join(xdgConfigDir(), "themes")
	if _, err := os.Stat(themesDir); os.IsNotExist(err) {
		return nil
	}

	files, err := filepath.Glob(filepath.Join(themesDir, "*.toml"))
	if err != nil {
		return nil
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var et ExternalTheme
		if err := toml.Unmarshal(data, &et); err != nil {
			continue
		}
		if et.Name == "" {
			stem := filepath.Base(file)
			stem = strings.TrimSuffix(stem, filepath.Ext(stem))
			et.Name = stem
		}
		if issues := ThemeContrastIssues(et.ToTheme()); len(issues) > 0 {
			lowContrast = append(lowContrast, fmt.Sprintf("%s (%s)", filepath.Base(file), strings.Join(issues, "; ")))
		}
	}
	return lowContrast
}

func xdgConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configHome, "pomogo")
}

// ResolveThemeName returns a concrete theme name, resolving "random" and "daily".
func ResolveThemeName(configured string) string {
	if configured == "random" {
		themes := List()
		if len(themes) == 0 {
			return "tokyo-night"
		}
		importTimeSeed := time.Now().UnixNano() + int64(os.Getpid())
		idx := int(importTimeSeed % int64(len(themes)))
		if idx < 0 {
			idx = -idx
		}
		return themes[idx]
	}

	if configured == "daily" {
		dateStr := time.Now().Format("2006-01-02")
		var hash int64
		for _, c := range dateStr {
			hash = hash*31 + int64(c)
		}
		themes := List()
		if len(themes) == 0 {
			return "tokyo-night"
		}
		idx := int(hash % int64(len(themes)))
		if idx < 0 {
			idx = -idx
		}
		return themes[idx]
	}

	return configured
}
