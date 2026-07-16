package render

import "testing"

func TestTextWidthHandlesWideGlyphs(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"Go", 2},
		{"🐹 Go", 5},
		{"界", 2},
	}

	for _, tt := range tests {
		if got := TextWidth(tt.text); got != tt.want {
			t.Errorf("TextWidth(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}

func TestPadRightUsesDisplayWidth(t *testing.T) {
	got := PadRight("🐹 Go", 7)
	if TextWidth(got) != 7 {
		t.Errorf("padded width = %d, want 7", TextWidth(got))
	}
}

func TestTruncateTextUsesDisplayWidth(t *testing.T) {
	got := TruncateText("🐹 Fawz Project", 8, "…")
	if TextWidth(got) > 8 {
		t.Errorf("truncated width = %d, want <= 8", TextWidth(got))
	}
	if got == "🐹 Fawz Project" {
		t.Error("expected text to be truncated")
	}
}
