package screens

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Ibnu-Afdel/pomogo/internal/theme"
)

// HelpBinding represents a key mapping to be shown on the help screen.
type HelpBinding struct {
	Keys        string
	Description string
}

// Help renders the help overlay modal screen.
func Help(width, height int, th *theme.Theme, bindings []HelpBinding) string {
	accent := lipgloss.Color(th.Accent.String())
	muted := lipgloss.Color(th.Muted.String())

	key := func(s string) string {
		return lipgloss.NewStyle().Foreground(accent).Bold(true).Render(s)
	}
	dim := func(s string) string {
		return lipgloss.NewStyle().Foreground(muted).Render(s)
	}

	var rows []string
	for _, b := range bindings {
		// Calculate dynamic padding to align descriptions
		padLen := 14 - len(b.Keys)
		if padLen < 1 {
			padLen = 1
		}
		rows = append(rows, key(b.Keys)+strings.Repeat(" ", padLen)+b.Description)
	}

	rows = append(rows, "")
	rows = append(rows, dim("Sessions auto-restore after an unexpected close."))
	rows = append(rows, dim("Press ? or Esc to close this overlay."))

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(1, 3).
		Render(strings.Join(rows, "\n"))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
