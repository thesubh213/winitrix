package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
)

func TestBuildLogoFrames(t *testing.T) {
	frames := buildLogoFrames(logoArt, 4)
	if len(frames) != 4 {
		t.Fatalf("got %d frames, want 4", len(frames))
	}

	for i, frame := range frames {
		if lines := strings.Split(frame, "\n"); len(lines) != len(logoArt) {
			t.Fatalf("frame %d has %d lines, want %d", i, len(lines), len(logoArt))
		}
		if !strings.Contains(frame, "\x1b[38;2;") {
			t.Fatalf("frame %d contains no truecolor sequences", i)
		}
	}

	if frames[0] == frames[1] {
		t.Fatal("consecutive frames should differ (animated phase)")
	}

	again := buildLogoFrames(logoArt, 4)
	for i := range frames {
		if frames[i] != again[i] {
			t.Fatalf("frame generation must be deterministic (frame %d)", i)
		}
	}
}

func TestLogoFrameSizeAndMemoryBudget(t *testing.T) {
	single := buildLogoFrames(logoArt, 1)[0]
	if len(single) > 12*1024 {
		t.Fatalf("single logo frame is %d bytes, expected well under 12KB", len(single))
	}

	f := newFX(newTheme(false), true)
	total := 0
	for _, frame := range f.logo {
		total += len(frame)
	}
	if total > 400*1024 {
		t.Fatalf("precomputed logo frames use %d bytes; budget is 400KB", total)
	}
}

func TestStaticModeHasSingleFrame(t *testing.T) {
	f := newFX(newTheme(false), false)
	if len(f.logo) != 1 {
		t.Fatalf("animation-free mode must bake a single frame, got %d", len(f.logo))
	}
}

func TestSpinnerAdvancesAndCycles(t *testing.T) {
	f := newFX(newTheme(false), true)

	f.frame = 0
	first := f.spinnerGlyph()
	f.frame = 2
	second := f.spinnerGlyph()
	if first == second {
		t.Fatal("spinner glyph should advance every other frame")
	}
	if first == "" || second == "" {
		t.Fatal("spinner glyph must not be empty")
	}

	// Must never panic or produce empty output while cycling.
	for frame := 0; frame < 500; frame++ {
		f.frame = frame
		if f.spinnerGlyph() == "" {
			t.Fatalf("empty spinner glyph at frame %d", frame)
		}
	}
}

func TestBarExactWidths(t *testing.T) {
	f := newFX(newTheme(false), true)

	for _, width := range []int{1, 2, 5, 10, 24, 60} {
		for _, frac := range []float64{-1, 0, 0.01, 0.37, 0.5, 0.99, 1, 2} {
			got := f.bar(frac, width, false)
			if w := ansi.StringWidth(got); w != width {
				t.Fatalf("bar(%v, %d) rendered %d cells, want %d", frac, width, w, width)
			}
		}
		if w := ansi.StringWidth(f.bar(0, width, true)); w != width {
			t.Fatalf("shimmer bar width = %d, want %d", w, width)
		}
	}
}

func TestBarClampsFractions(t *testing.T) {
	f := newFX(newTheme(false), true)

	if f.bar(2, 10, false) != f.bar(1, 10, false) {
		t.Fatal("fractions above 1 must clamp")
	}
	if f.bar(-1, 10, false) != f.bar(0, 10, false) {
		t.Fatal("negative fractions must clamp to 0")
	}
	if f.bar(0, 10, false) == f.bar(1, 10, false) {
		t.Fatal("empty and full bars must differ")
	}
}

func TestBarShimmerMoves(t *testing.T) {
	f := newFX(newTheme(false), true)

	f.frame = 0
	a := f.bar(0, 20, true)
	f.frame = 3
	b := f.bar(0, 20, true)
	if a == b {
		t.Fatal("shimmer must move with the frame counter")
	}
}

func TestRuleWidthsAndCache(t *testing.T) {
	f := newFX(newTheme(false), true)

	for _, w := range []int{1, 40, 118} {
		if got := ansi.StringWidth(f.rule(w)); got != w {
			t.Fatalf("rule(%d) width = %d", w, got)
		}
	}
	first := f.rule(40)
	if f.rule(40) != first {
		t.Fatal("rule must be cached deterministically")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, "0s"},
		{900 * time.Millisecond, "1s"},
		{28 * time.Second, "28s"},
		{62 * time.Second, "1m 02s"},
		{time.Hour, "60m 00s"},
		{-5 * time.Second, "0s"},
	}

	for _, tt := range tests {
		if got := formatDuration(tt.in); got != tt.want {
			t.Errorf("formatDuration(%s) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestPulseColors(t *testing.T) {
	colors := pulseColors(SecondaryColor, PrimaryColor, pulseSteps)
	if len(colors) != pulseSteps {
		t.Fatalf("got %d colors, want %d", len(colors), pulseSteps)
	}
	for i, c := range colors {
		if c == nil {
			t.Fatalf("pulse color %d is nil", i)
		}
	}
}
