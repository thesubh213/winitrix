package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestTruncateToWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
	}{
		{"plain short", "hello", 10},
		{"plain exact", "hello", 5},
		{"plain long", "hello world", 4},
		{"ansi styled", lipgloss.NewStyle().Bold(true).Foreground(PrimaryColor).Render("styled text"), 6},
		{"wide runes", "日本語テキスト", 5},
		{"emoji", "build 🚀 complete", 8},
		{"zero width", "anything", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.width)
			if w := ansi.StringWidth(got); w > tt.width {
				t.Fatalf("truncate(%q, %d) = %q (width %d), exceeds limit", tt.input, tt.width, got, w)
			}
		})
	}
}

func TestPadHelpers(t *testing.T) {
	styled := lipgloss.NewStyle().Bold(true).Render("abc")

	if got := padRight(styled, 6); ansi.StringWidth(got) != 6 {
		t.Fatalf("padRight width = %d, want 6", ansi.StringWidth(got))
	}
	if got := padLeft("abc", 6); ansi.StringWidth(got) != 6 || !strings.HasPrefix(got, "   ") {
		t.Fatalf("padLeft = %q", got)
	}
	// Already-wide strings are returned untouched.
	long := "abcdef"
	if got := padRight(long, 3); got != long {
		t.Fatalf("padRight must not truncate: %q", got)
	}
}

func TestWindowRange(t *testing.T) {
	tests := []struct {
		total, cursor, maxRows int
		wantStart, wantEnd     int
	}{
		{5, 0, 10, 0, 5},
		{10, 0, 4, 0, 4},
		{10, 5, 4, 3, 7},
		{10, 9, 4, 6, 10},
		{10, 0, 3, 0, 3},
		{10, 5, 0, 0, 10},
		{3, 99, 2, 1, 3},
	}

	for _, tt := range tests {
		start, end := windowRange(tt.total, tt.cursor, tt.maxRows)
		if start != tt.wantStart || end != tt.wantEnd {
			t.Errorf("windowRange(%d, %d, %d) = (%d, %d), want (%d, %d)",
				tt.total, tt.cursor, tt.maxRows, start, end, tt.wantStart, tt.wantEnd)
		}
	}
}

func TestRenderPanelWidthConsistency(t *testing.T) {
	rows := []string{
		"short",
		"Installing: some-very-long-package-name-with-version [1.2.3 42%] extra text to force truncation",
		"",
	}

	panel := renderPanel(rows, 40, PrimaryColor, unicodeGlyphs())
	lines := strings.Split(panel, "\n")
	if len(lines) != len(rows)+2 {
		t.Fatalf("panel has %d lines, want %d", len(lines), len(rows)+2)
	}
	for _, line := range lines {
		if w := ansi.StringWidth(line); w != 40 {
			t.Fatalf("panel line width = %d, want 40: %q", w, line)
		}
	}
}

func TestRenderPanelNarrowWidth(t *testing.T) {
	panel := renderPanel([]string{"content that is quite long and should be truncated"}, 20, PrimaryColor, unicodeGlyphs())
	for _, line := range strings.Split(panel, "\n") {
		if w := ansi.StringWidth(line); w != 20 {
			t.Fatalf("narrow panel line width = %d, want 20: %q", w, line)
		}
	}
}

func TestRenderPanelASCII(t *testing.T) {
	panel := renderPanel([]string{"row"}, 12, PrimaryColor, asciiGlyphs())
	plain := ansi.Strip(panel)
	for _, r := range plain {
		if r > 126 {
			t.Fatalf("ascii panel contains non-ascii rune %q in %q", r, plain)
		}
	}
}

func TestHighlightRowFillsWidth(t *testing.T) {
	row := highlightRow("cursor row", 40)
	if w := ansi.StringWidth(row); w != innerWidth(40) {
		t.Fatalf("row width = %d, want %d", w, innerWidth(40))
	}
}

func TestKeyHintsRendering(t *testing.T) {
	hints := keyHints("space", "toggle", "a", "all")
	plain := ansi.Strip(hints)
	if !strings.Contains(plain, "[space] toggle") || !strings.Contains(plain, "[a] all") {
		t.Fatalf("keyHints output = %q", plain)
	}
}

func TestStatusGlyph(t *testing.T) {
	g := unicodeGlyphs()
	if ansi.Strip(statusGlyph(g, rowDone)) == "" {
		t.Fatal("done glyph is empty")
	}
	if statusGlyph(g, rowActive) != "" {
		t.Fatal("active rows use the animated spinner, not a static glyph")
	}
	if !strings.Contains(ansi.Strip(statusGlyph(g, rowPending)), g.pending) {
		t.Fatal("pending glyph mismatch")
	}
}
