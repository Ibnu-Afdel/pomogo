package render

import (
	"strings"
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/Ibnu-Afdel/pomogo/internal/timer"
	"github.com/charmbracelet/lipgloss"
)

func TestLayoutVisualQualityGate(t *testing.T) {
	th := theme.TokyoNight()
	sizes := []Frame{
		{Width: 80, Height: 24},
		{Width: 120, Height: 32},
	}
	states := map[string]DisplayState{
		"quick": {
			ModeLabel:        "Building",
			Project:          "Fawz",
			Task:             "build auth",
			PhaseKind:        timer.PhaseWork,
			SegmentRemaining: 18*time.Minute + 32*time.Second,
			Progress:         0.36,
			SegmentIndex:     2,
			SegmentCount:     4,
			Running:          true,
			StatusMessage:    "focus · break in 19m",
			ThemeName:        "tokyo-night",
		},
		"deep": {
			ModeLabel:        "Building",
			Project:          "Fawz",
			Task:             "build auth",
			PhaseKind:        timer.PhaseWork,
			SegmentRemaining: 18*time.Minute + 32*time.Second,
			BlockRemaining:   2*time.Hour + 12*time.Minute + 9*time.Second,
			Progress:         0.42,
			Running:          true,
			StatusMessage:    "focus · break in 19m",
			ThemeName:        "tokyo-night",
		},
	}

	for layoutName, spec := range Registry {
		for _, frame := range sizes {
			if frame.Width < spec.MinWidth || frame.Height < spec.MinHeight {
				continue
			}
			for stateName, ds := range states {
				t.Run(layoutName+"/"+stateName+"/"+frameName(frame), func(t *testing.T) {
					assertFrameFits(t, spec.Layout(ds, th, frame), frame)
					ds.Zen = true
					assertFrameFits(t, spec.Layout(ds, th, frame), frame)
				})
			}
		}
	}
}

func assertFrameFits(t *testing.T, got string, frame Frame) {
	t.Helper()
	if strings.TrimSpace(stripANSI(got)) == "" {
		t.Fatal("rendered frame is empty")
	}
	lines := strings.Split(got, "\n")
	if len(lines) > frame.Height {
		t.Fatalf("rendered frame height %d exceeds terminal height %d", len(lines), frame.Height)
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > frame.Width {
			t.Fatalf("line %d width %d exceeds terminal width %d: %q", i, w, frame.Width, line)
		}
	}
}

func frameName(f Frame) string {
	return strings.Join([]string{itoa(f.Width), "x", itoa(f.Height)}, "")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func stripANSI(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			i++
			for i < len(s) {
				if s[i] >= 0x40 && s[i] <= 0x7e {
					break
				}
				i++
			}
			continue
		}
		out.WriteByte(s[i])
	}
	return out.String()
}
