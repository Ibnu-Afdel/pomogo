package render

import (
	"strings"
	"testing"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

// TestAmbientDeterminism verifies that rendering with the same inputs produces identical outputs.
func TestAmbientDeterminism(t *testing.T) {
	th := theme.TokyoNight()
	f := Frame{Width: 80, Height: 24}
	content := "Line 1\nLine 2"

	effects := []string{"stars", "snow", "rain"}
	for _, eff := range effects {
		t.Run(eff, func(t *testing.T) {
			res1 := RenderAmbient(eff, 12345, f, th, content)
			res2 := RenderAmbient(eff, 12345, f, th, content)

			if res1 != res2 {
				t.Errorf("ambient effect %s is not deterministic", eff)
			}
		})
	}
}

func TestAmbientShowsThroughWhitespace(t *testing.T) {
	th := theme.TokyoNight()
	f := Frame{Width: 40, Height: 10}
	var lines []string
	for i := 0; i < f.Height; i++ {
		lines = append(lines, strings.Repeat(" ", f.Width))
	}
	content := strings.Join(lines, "\n")

	got := RenderAmbient("stars", 1, f, th, content)
	if got == content {
		t.Fatal("ambient output stayed blank; whitespace should reveal particles")
	}
	if !strings.ContainsAny(got, ".*+") {
		t.Fatalf("ambient output does not contain visible star particles: %q", got)
	}
}
