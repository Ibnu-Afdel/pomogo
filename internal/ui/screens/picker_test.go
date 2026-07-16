package screens

import (
	"strings"
	"testing"
	"time"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

func TestDurationPickerShowsDefaultAndRhythm(t *testing.T) {
	got := DurationPicker(80, 24, theme.TokyoNight(), 1, 2*time.Hour, 25*time.Minute, 5*time.Minute, 15*time.Minute, 4)
	if !strings.Contains(got, "2h block · 25m/5m/15m rhythm every 4") {
		t.Fatalf("picker missing rhythm line: %q", got)
	}
	if !strings.Contains(got, "120 min (2 hours)  default") {
		t.Fatalf("picker missing default marker: %q", got)
	}
	if !strings.Contains(got, "240 min (4 hours)") {
		t.Fatalf("picker missing minute-based 4h option: %q", got)
	}
}
