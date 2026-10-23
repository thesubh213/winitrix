package tui

import (
	"image/color"
	"testing"
)

func TestGradientSeqs(t *testing.T) {
	seqs := gradientSeqs([]color.Color{gradDeep, gradSky, gradAqua}, gradientSteps)

	if len(seqs) != gradientSteps {
		t.Fatalf("got %d sequences, want %d", len(seqs), gradientSteps)
	}
	for i, seq := range seqs {
		if len(seq) < 10 || seq[0] != 0x1b {
			t.Fatalf("sequence %d is not an ANSI escape: %q", i, seq)
		}
	}
	if seqs[0] == seqs[len(seqs)-1] {
		t.Fatal("gradient endpoints should differ")
	}
}

func TestGlyphSets(t *testing.T) {
	uni := unicodeGlyphs()
	ascii := asciiGlyphs()

	for name, g := range map[string]glyphSet{"unicode": uni, "ascii": ascii} {
		if g.check == "" || g.cross == "" || g.barFull == "" || g.barEmpty == "" {
			t.Fatalf("%s glyph set has empty entries", name)
		}
		if len(g.spinner) == 0 {
			t.Fatalf("%s glyph set has no spinner frames", name)
		}
	}

	asciiGlyphsAll := append([]string{}, ascii.spinner...)
	asciiGlyphsAll = append(asciiGlyphsAll, ascii.check, ascii.cross, ascii.cursor, ascii.markOn, ascii.markOff, ascii.barFull, ascii.barEmpty, ascii.rule)
	for _, glyph := range asciiGlyphsAll {
		for _, r := range glyph {
			if r > 126 {
				t.Fatalf("ascii glyph %q contains non-ascii rune %q", glyph, r)
			}
		}
	}
}

func TestNewTheme(t *testing.T) {
	if th := newTheme(false); th.glyphs.boxV != "│" {
		t.Fatal("expected unicode theme")
	}
	if th := newTheme(true); th.glyphs.boxV != "|" {
		t.Fatal("expected ascii theme")
	}
}

func TestStateColorsAreDistinct(t *testing.T) {
	type rgba struct{ r, g, b uint32 }
	seen := map[rgba]state{}
	for _, s := range []state{stateScanning, stateUpdating, stateErrorPrompt, stateConfirmCancel, stateDone} {
		r, g, b, _ := stateColor(s).RGBA()
		key := rgba{r, g, b}
		if prev, ok := seen[key]; ok {
			t.Fatalf("states %v and %v share a color", prev, s)
		}
		seen[key] = s
	}
}

func TestLogoArtShape(t *testing.T) {
	if len(logoArt) < 5 {
		t.Fatalf("logo art has %d lines, expected the full wordmark", len(logoArt))
	}

	width := 0
	for _, line := range logoArt {
		if n := len([]rune(line)); n > width {
			width = n
		}
	}
	if width == 0 {
		t.Fatal("logo art is empty")
	}
}
