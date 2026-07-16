package render

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// bigDigits[d][row] is row `row` of digit `d`. Each row is exactly 4 visible chars.
var bigDigits = [10][5]string{
	{" ██ ", "█  █", "█  █", "█  █", " ██ "}, // 0
	{"  █ ", "  █ ", "  █ ", "  █ ", "  █ "}, // 1
	{" ██ ", "   █", " ██ ", "█   ", "████"}, // 2
	{" ██ ", "   █", " ██ ", "   █", " ██ "}, // 3
	{"█  █", "█  █", "████", "   █", "   █"}, // 4
	{"████", "█   ", "███ ", "   █", "███ "}, // 5
	{" ██ ", "█   ", "███ ", "█  █", " ██ "}, // 6
	{"████", "  █ ", " █  ", " █  ", " █  "}, // 7
	{" ██ ", "█  █", " ██ ", "█  █", " ██ "}, // 8
	{" ██ ", "█  █", " ███", "   █", " ██ "}, // 9
}

// bigColon[row] is row `row` of the ":" separator. Each row is exactly 2 visible chars.
var bigColon = [5]string{"  ", " ●", "  ", " ●", "  "}

// BigClockRows returns 5 rows of large ASCII-art digits for a time string (e.g. "MM:SS" or "H:MM:SS"),
// styled with the given color.
func BigClockRows(timeStr string, color lipgloss.Color) []string {
	style := lipgloss.NewStyle().Foreground(color).Bold(true)
	var rows [5]strings.Builder
	first := true
	for _, ch := range timeStr {
		if !first {
			for i := range rows {
				rows[i].WriteByte(' ')
			}
		}
		first = false
		switch {
		case ch >= '0' && ch <= '9':
			d := bigDigits[ch-'0']
			for i, seg := range d {
				rows[i].WriteString(seg)
			}
		case ch == ':':
			for i, seg := range bigColon {
				rows[i].WriteString(seg)
			}
		}
	}
	result := make([]string, 5)
	for i, sb := range rows {
		result[i] = style.Render(sb.String())
	}
	return result
}
