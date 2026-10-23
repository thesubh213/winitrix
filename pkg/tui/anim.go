package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	frameInterval  = 50 * time.Millisecond // ~20 fps while animating
	idleInterval   = time.Second           // clock-only refresh
	logoFrameCount = 32
	gradientSteps  = 32
	spinnerPulses  = 20
	pulseSteps     = 9
	shimmerWidth   = 6
)

// animTickMsg drives every animation frame.
type animTickMsg struct{}

func tickCmd(animating bool) tea.Cmd {
	d := idleInterval
	if animating {
		d = frameInterval
	}
	return tea.Tick(d, func(time.Time) tea.Msg { return animTickMsg{} })
}

var logoArt = []string{
	" ██╗    ██╗██╗███╗   ██╗██╗████████╗██████╗ ██╗██╗  ██╗",
	" ██║    ██║██║████╗  ██║██║╚══██╔══╝██╔══██╗██║╚██╗██╔╝",
	" ██║ █╗ ██║██║██╔██╗ ██║██║   ██║   ██████╔╝██║ ╚███╔╝ ",
	" ██║███╗██║██║██║╚██╗██║██║   ██║   ██╔══██╗██║ ██╔██╗ ",
	" ╚███╔███╔╝██║██║ ╚████║██║   ██║   ██║  ██║██║██╔╝ ██╗",
	"  ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝",
}

// partialBlocks are the eighth-block glyphs used for sub-cell progress.
var partialBlocks = [8]string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"}

// fgSeq renders a 24-bit foreground escape sequence. Full-fidelity ANSI is
// safe to embed: Bubble Tea v2 downsamples colors to the terminal profile
// when it writes frames.
func fgSeq(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r>>8, g>>8, b>>8)
}

// gradientSeqs returns n precomputed foreground escape sequences along a
// CIELAB gradient through the provided stops.
func gradientSeqs(stops []color.Color, n int) []string {
	colors := lipgloss.Blend1D(n, stops...)
	seqs := make([]string, len(colors))
	for i, c := range colors {
		seqs[i] = fgSeq(c)
	}
	return seqs
}

// buildLogoFrames bakes the animated gradient wordmark into a frame table so
// the render path is a single string lookup per frame.
func buildLogoFrames(art []string, frames int) []string {
	if frames < 1 {
		frames = 1
	}

	width := 0
	for _, line := range art {
		if n := len([]rune(line)); n > width {
			width = n
		}
	}
	if width == 0 {
		return []string{""}
	}

	seqs := gradientSeqs([]color.Color{gradDeep, gradSky, gradAqua}, gradientSteps)

	out := make([]string, frames)
	for f := 0; f < frames; f++ {
		phase := float64(f) / float64(frames)
		var b strings.Builder
		for y, line := range art {
			runes := []rune(line)
			prev := ""
			for x := 0; x < width; x++ {
				var r = ' '
				if x < len(runes) {
					r = runes[x]
				}
				if r == ' ' {
					b.WriteRune(' ')
					prev = ""
					continue
				}
				t := math.Mod(float64(x)/float64(width)+phase+float64(y)*0.05, 1)
				if t < 0 {
					t += 1
				}
				seq := seqs[int(t*float64(gradientSteps-1))]
				if seq != prev {
					b.WriteString(seq)
					prev = seq
				}
				b.WriteRune(r)
			}
			b.WriteString("\x1b[0m")
			if y < len(art)-1 {
				b.WriteString("\n")
			}
		}
		out[f] = b.String()
	}
	return out
}

func buildSpinnerTable(g glyphSet, pulses int) [][]string {
	pulse := lipgloss.Blend1D(pulses, AccentColor, PrimaryColor)
	table := make([][]string, pulses)
	for p := range table {
		style := lipgloss.NewStyle().Foreground(pulse[p]).Bold(true)
		row := make([]string, len(g.spinner))
		for i, glyph := range g.spinner {
			row[i] = style.Render(glyph)
		}
		table[p] = row
	}
	return table
}

// barCache holds width-specific pre-rendered bar pieces.
type barCache struct {
	seqs   []string // gradient foreground sequence per cell
	bright string   // shimmer foreground sequence
	empty  string   // styled empty cell glyph
	full   string   // full cell glyph
}

// fx is the shared animation/render engine. It is intentionally pointer-held
// in the model so value copies stay cheap.
type fx struct {
	glyphs  glyphSet
	frame   int
	logo    []string
	spinner [][]string
	bars    map[int]*barCache
	rules   map[int]string
}

func newFX(t theme, animations bool) *fx {
	frames := 1
	if animations {
		frames = logoFrameCount
	}
	return &fx{
		glyphs:  t.glyphs,
		logo:    buildLogoFrames(logoArt, frames),
		spinner: buildSpinnerTable(t.glyphs, spinnerPulses),
		bars:    make(map[int]*barCache),
		rules:   make(map[int]string),
	}
}

// rule renders a gradient horizontal rule, cached per width.
func (s *fx) rule(width int) string {
	if width < 1 {
		width = 1
	}
	if r, ok := s.rules[width]; ok {
		return r
	}

	seqs := gradientSeqs([]color.Color{gradDeep, gradSky, gradAqua}, width)
	var b strings.Builder
	for i := 0; i < width; i++ {
		b.WriteString(seqs[i])
		b.WriteString("─")
	}
	b.WriteString("\x1b[0m")
	out := b.String()
	s.rules[width] = out
	return out
}

func (s *fx) logoFrame() string {
	if len(s.logo) == 0 {
		return ""
	}
	return s.logo[s.frame%len(s.logo)]
}

func (s *fx) spinnerGlyph() string {
	if len(s.spinner) == 0 {
		return ""
	}
	row := s.spinner[s.frame%len(s.spinner)]
	if len(row) == 0 {
		return ""
	}
	// Advance the spinner glyph at half the frame rate so a full rotation
	// takes about a second.
	return row[(s.frame/2)%len(row)]
}

func (s *fx) barCacheFor(width int) *barCache {
	if width < 1 {
		width = 1
	}
	if c, ok := s.bars[width]; ok {
		return c
	}

	seqs := gradientSeqs([]color.Color{gradDeep, gradSky, gradAqua, SuccessColor}, width)
	c := &barCache{
		seqs:   seqs,
		bright: fgSeq(lipgloss.Color("#E6FFFF")),
		empty:  lipgloss.NewStyle().Foreground(MutedColor).Render(s.glyphs.barEmpty),
		full:   s.glyphs.barFull,
	}
	s.bars[width] = c
	return c
}

// bar renders a progress bar. When indeterminate is true a shimmer band
// sweeps across the empty track instead of showing a fill.
func (s *fx) bar(fraction float64, width int, indeterminate bool) string {
	if width < 1 {
		return ""
	}
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}

	c := s.barCacheFor(width)
	var b strings.Builder

	if indeterminate {
		bandPos := s.frame % (width + shimmerWidth)
		for i := 0; i < width; i++ {
			if i >= bandPos-shimmerWidth && i <= bandPos {
				b.WriteString(c.bright)
				b.WriteString(c.full)
			} else {
				b.WriteString(c.empty)
			}
		}
		return b.String()
	}

	exact := fraction * float64(width)
	filled := int(exact)
	if fraction > 0 && filled == 0 {
		filled = 0
	}

	for i := 0; i < filled && i < width; i++ {
		b.WriteString(c.seqs[i])
		b.WriteString(c.full)
	}

	if filled < width {
		remainder := exact - float64(filled)
		if idx := int(remainder * 8); idx > 0 {
			if idx > 7 {
				idx = 7
			}
			glyph := partialBlocks[idx]
			if s.glyphs.barFull != "█" {
				glyph = s.glyphs.barFull
			}
			b.WriteString(c.seqs[filled])
			b.WriteString(glyph)
			filled++
		}
	}

	for i := filled; i < width; i++ {
		b.WriteString(c.empty)
	}

	return b.String()
}

// pulseColors interpolates border colors for state-transition pulses.
func pulseColors(from, to color.Color, steps int) []color.Color {
	if from == nil {
		from = to
	}
	return lipgloss.Blend1D(steps, from, to)
}

// formatDuration renders durations compactly ("12s", "1m 04s").
func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()+0.5))
	}
	m := int(d / time.Minute)
	s := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm %02ds", m, s)
}
