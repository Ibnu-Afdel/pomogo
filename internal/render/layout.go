// Package render provides pure rendering layouts and widgets for PomoGo.
package render

import (
	"os"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
)

// Frame defines the dimensions of the terminal frame.
type Frame struct {
	Width  int
	Height int
}

// DisplayState represents a snapshot of the UI state to be rendered.
type DisplayState struct {
	ModeLabel        string
	Project          string
	Task             string
	PhaseKind        timer.SessionPhase
	SegmentRemaining time.Duration
	BlockRemaining   time.Duration
	Progress         float64 // progress from 0.0 to 1.0
	SegmentIndex     int     // e.g. completed sessions in this cycle
	SegmentCount     int     // e.g. total sessions before long break
	Paused           bool
	Running          bool
	Idle             bool
	StatusMessage    string
	HintsVisibility  bool
	ThemeName        string
	LayoutName       string
	Zen              bool
	GitBranch        string
	TmuxSession      string
}

// Layout is a pure function that takes a DisplayState, a Theme, and a Frame, and returns the rendered string.
type Layout func(ds DisplayState, th *theme.Theme, f Frame) string

// LayoutSpec defines layout limits.
type LayoutSpec struct {
	Layout    Layout
	MinWidth  int
	MinHeight int
}

// Registry maps layout names to LayoutSpecs.
var Registry = map[string]LayoutSpec{}

// ResolveLayout determines the best layout to fit the terminal dimensions.
func ResolveLayout(name string, width, height int) (string, Layout) {
	// If requested layout fits, use it
	if spec, exists := Registry[name]; exists {
		if width >= spec.MinWidth && height >= spec.MinHeight {
			return name, spec.Layout
		}
	}

	// Otherwise, find the first layout that fits in preference order
	order := []string{"tinybar", "minimal", "compact", "focus-stack", "centered", "classic", "dashboard", "monolith", "retro", "terminal-rice", "command-center"}
	for _, lName := range order {
		spec, exists := Registry[lName]
		if exists && width >= spec.MinWidth && height >= spec.MinHeight {
			return lName, spec.Layout
		}
	}

	// Fallback to minimal
	return "minimal", Registry["minimal"].Layout
}

// ResolveLayoutName resolves "random" and "daily" to a concrete layout name.
func ResolveLayoutName(configured string) string {
	layouts := []string{"classic", "minimal", "centered", "compact", "retro", "dashboard", "monolith", "tinybar", "terminal-rice", "focus-stack", "command-center"}
	if configured == "random" {
		importTimeSeed := time.Now().UnixNano() + int64(os.Getpid())
		idx := int(importTimeSeed % int64(len(layouts)))
		if idx < 0 {
			idx = -idx
		}
		return layouts[idx]
	}
	if configured == "daily" {
		dateStr := time.Now().Format("2006-01-02")
		var hash int64
		for _, c := range dateStr {
			hash = hash*31 + int64(c)
		}
		idx := int(hash % int64(len(layouts)))
		if idx < 0 {
			idx = -idx
		}
		return layouts[idx]
	}
	if configured == "" {
		return "classic"
	}
	return configured
}
