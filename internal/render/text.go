package render

import "github.com/mattn/go-runewidth"

// TextWidth returns the display width of raw terminal text.
func TextWidth(s string) int {
	return runewidth.StringWidth(s)
}

// PadRight pads raw terminal text to a target display width.
func PadRight(s string, width int) string {
	w := TextWidth(s)
	if w >= width {
		return s
	}
	return s + spaces(width-w)
}

// TruncateText trims raw terminal text to a target display width.
func TruncateText(s string, width int, tail string) string {
	if width <= 0 {
		return ""
	}
	if TextWidth(s) <= width {
		return s
	}
	tailW := TextWidth(tail)
	if tailW >= width {
		return runewidth.Truncate(tail, width, "")
	}
	return runewidth.Truncate(s, width-tailW, "") + tail
}

func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}
