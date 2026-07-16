package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// SessionDots renders a series of ● and ○ indicating progress through the current work cycle.
func SessionDots(completed, total int, filled, empty lipgloss.Color) string {
	on := lipgloss.NewStyle().Foreground(filled)
	off := lipgloss.NewStyle().Foreground(empty)
	var sb strings.Builder
	for i := 0; i < total; i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		if i < completed {
			sb.WriteString(on.Render("●"))
		} else {
			sb.WriteString(off.Render("○"))
		}
	}
	return sb.String()
}

// ProgressBar renders a ━─ style bar of exactly `width` visible characters.
func ProgressBar(progress float64, width int, filled, empty lipgloss.Color) string {
	n := int(float64(width) * progress)
	if n < 0 {
		n = 0
	}
	if n > width {
		n = width
	}
	f := lipgloss.NewStyle().Foreground(filled).Render(strings.Repeat("━", n))
	e := lipgloss.NewStyle().Foreground(empty).Render(strings.Repeat("─", width-n))
	return f + e
}
