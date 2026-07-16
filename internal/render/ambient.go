package render

import (
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Ibnu-Afdel/pomogo/internal/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// hash is a fast, deterministic pseudo-random hash function with 0 allocations.
func hash(x, y, seed int) uint32 {
	h := uint32(x)*37 + uint32(y)*17 + uint32(seed)
	h = (h ^ 61) ^ (h >> 16)
	h *= 9
	h = (h ^ (h >> 4)) * 0x27d4eb2d
	h = h ^ (h >> 15)
	return h
}

// RenderAmbient renders the ambient particle background and overlays the timer layout.
func RenderAmbient(effect string, tickCount int, f Frame, th *theme.Theme, layoutContent string) string {
	if effect == "none" || effect == "" {
		return layoutContent
	}

	// Pre-render styled particle characters to avoid Lip Gloss allocations in loops
	starDim := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Ambient.String())).Render(".")
	starMed := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted.String())).Render("+")
	starBright := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted.String())).Bold(true).Render("*")

	snow1 := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Ambient.String())).Render(".")
	snow2 := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted.String())).Render("*")

	rain1 := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Ambient.String())).Render("│")
	rain2 := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Ambient.String())).Render("/")

	space := " "

	// Create background grid of styled strings
	bgGrid := make([][]string, f.Height)
	for y := 0; y < f.Height; y++ {
		bgGrid[y] = make([]string, f.Width)
		for x := 0; x < f.Width; x++ {
			char := space
			switch effect {
			case "stars":
				if hash(x, y, 42)%100 < 3 {
					blink := (hash(x, y, 12) + uint32(tickCount)) % 4
					switch blink {
					case 0:
						char = starDim
					case 1:
						char = starMed
					case 2:
						char = starBright
					default:
						char = space
					}
				}
			case "snow":
				offset := hash(x, 0, 77)
				row := (offset + uint32(tickCount)) % uint32(f.Height)
				if y == int(row) && hash(x, 0, 88)%4 == 0 {
					if hash(x, y, 99)%2 == 0 {
						char = snow2
					} else {
						char = snow1
					}
				}
			case "rain":
				offset := hash(x, 0, 11)
				row := (offset + uint32(tickCount*2)) % uint32(f.Height)
				if y == int(row) && hash(x, 0, 22)%3 == 0 {
					if hash(x, y, 33)%2 == 0 {
						char = rain2
					} else {
						char = rain1
					}
				}
			}
			bgGrid[y][x] = char
		}
	}

	// Split layout content into lines and compute size
	contentLines := strings.Split(layoutContent, "\n")
	if len(contentLines) > 0 && contentLines[len(contentLines)-1] == "" {
		contentLines = contentLines[:len(contentLines)-1]
	}

	contentHeight := len(contentLines)
	contentWidth := 0
	for _, line := range contentLines {
		w := lipgloss.Width(line)
		if w > contentWidth {
			contentWidth = w
		}
	}

	startY := (f.Height - contentHeight) / 2
	startX := (f.Width - contentWidth) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Render final composite frame
	merged := make([]string, f.Height)
	for y := 0; y < f.Height; y++ {
		if y >= startY && y < startY+contentHeight {
			lineIdx := y - startY
			contentLine := contentLines[lineIdx]

			overlayWidth := contentWidth
			if startX+overlayWidth > f.Width {
				overlayWidth = f.Width - startX
			}
			padded := lipgloss.NewStyle().Width(overlayWidth).Render(contentLine)
			leftBg := strings.Join(bgGrid[y][:startX], "")
			overlay := overlayLine(bgGrid[y][startX:startX+overlayWidth], padded)
			rightBg := strings.Join(bgGrid[y][startX+overlayWidth:], "")

			merged[y] = leftBg + overlay + rightBg
		} else {
			merged[y] = strings.Join(bgGrid[y], "")
		}
	}

	return strings.Join(merged, "\n")
}

func overlayLine(bg []string, content string) string {
	var out strings.Builder
	col := 0
	for i := 0; i < len(content); {
		if content[i] == '\x1b' {
			end := i + 1
			for end < len(content) {
				b := content[end]
				end++
				if b >= 0x40 && b <= 0x7e {
					break
				}
			}
			out.WriteString(content[i:end])
			i = end
			continue
		}

		r, size := utf8.DecodeRuneInString(content[i:])
		if r == utf8.RuneError && size == 0 {
			break
		}
		w := runewidth.RuneWidth(r)
		if r == ' ' && col < len(bg) {
			out.WriteString(bg[col])
		} else {
			out.WriteString(content[i : i+size])
		}
		if w > 0 {
			col += w
		}
		i += size
	}
	return out.String()
}

// ResolveEffectsName resolves "random" to a concrete effect name.
func ResolveEffectsName(configured string) string {
	effects := []string{"none", "stars", "snow", "rain"}
	if configured == "random" {
		importTimeSeed := time.Now().UnixNano() + int64(os.Getpid())
		idx := int(importTimeSeed % int64(len(effects)))
		if idx < 0 {
			idx = -idx
		}
		return effects[idx]
	}
	if configured == "" {
		return "none"
	}
	return configured
}
