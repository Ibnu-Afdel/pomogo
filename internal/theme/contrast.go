package theme

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ThemeContrastIssues returns readability issues for the core foreground roles.
func ThemeContrastIssues(th *Theme) []string {
	if th == nil {
		return []string{"theme is nil"}
	}
	checks := []struct {
		name      string
		fg        Color
		bg        Color
		min       float64
		essential bool
	}{
		{"text/background", th.Text, th.Background, 4.5, true},
		{"muted/background", th.Muted, th.Background, 2.0, false},
		{"accent/background", th.Accent, th.Background, 3.0, false},
	}

	var issues []string
	for _, c := range checks {
		ratio, err := ContrastRatio(c.fg, c.bg)
		if err != nil {
			issues = append(issues, fmt.Sprintf("%s invalid color: %v", c.name, err))
			continue
		}
		if ratio < c.min {
			issues = append(issues, fmt.Sprintf("%s contrast %.2f below %.1f", c.name, ratio, c.min))
		}
	}
	return issues
}

// ContrastRatio returns the WCAG contrast ratio between two hex colors.
func ContrastRatio(fg, bg Color) (float64, error) {
	fr, fgG, fb, err := parseHexColor(fg)
	if err != nil {
		return 0, err
	}
	br, bgG, bb, err := parseHexColor(bg)
	if err != nil {
		return 0, err
	}
	l1 := relativeLuminance(fr, fgG, fb)
	l2 := relativeLuminance(br, bgG, bb)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05), nil
}

func parseHexColor(c Color) (float64, float64, float64, error) {
	s := strings.TrimPrefix(strings.TrimSpace(string(c)), "#")
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("expected #rrggbb, got %q", c)
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse %q: %w", c, err)
	}
	r := float64((v>>16)&0xff) / 255
	g := float64((v>>8)&0xff) / 255
	b := float64(v&0xff) / 255
	return r, g, b, nil
}

func relativeLuminance(r, g, b float64) float64 {
	return 0.2126*linearRGB(r) + 0.7152*linearRGB(g) + 0.0722*linearRGB(b)
}

func linearRGB(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}
