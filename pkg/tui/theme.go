package tui

import (
	"image/color"
	"io"
	"strings"

	"charm.land/lipgloss/v2"
)

// ---------------------------------------------------------------------------
// Palette — curated deep-ocean theme
// ---------------------------------------------------------------------------

var (
	PrimaryColor   color.Color = lipgloss.Color("#00BFFF") // deep sky blue
	SecondaryColor color.Color = lipgloss.Color("#1E90FF") // dodger blue
	AccentColor    color.Color = lipgloss.Color("#5B9BD5") // steel blue
	SuccessColor   color.Color = lipgloss.Color("#00FF7F") // spring green
	ErrorColor     color.Color = lipgloss.Color("#FF4500") // orange red
	WarningColor   color.Color = lipgloss.Color("#FFD700") // gold
	SubtleColor    color.Color = lipgloss.Color("#696969") // dim gray
	MutedColor     color.Color = lipgloss.Color("#4A4A4A") // muted gray
	BgColor        color.Color = lipgloss.Color("#0F0F0F")

	// Gradient stops used by animated elements.
	gradDeep = lipgloss.Color("#0057FF")
	gradSky  = lipgloss.Color("#00BFFF")
	gradAqua = lipgloss.Color("#00FFF0")

	// Selection highlight background for the active row.
	highlightBg = lipgloss.Color("#123A5E")
)

// ---------------------------------------------------------------------------
// Style presets
// ---------------------------------------------------------------------------

var (
	SuccessStyle   = lipgloss.NewStyle().Foreground(SuccessColor).Bold(true)
	ErrorStyle     = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)
	WarningStyle   = lipgloss.NewStyle().Foreground(WarningColor).Bold(true)
	SubtleStyle    = lipgloss.NewStyle().Foreground(SubtleColor)
	MutedStyle     = lipgloss.NewStyle().Foreground(MutedColor)
	HighlightStyle = lipgloss.NewStyle().Foreground(PrimaryColor).Bold(true)
	AccentStyle    = lipgloss.NewStyle().Foreground(AccentColor)
	BoldStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	DimBoldStyle   = lipgloss.NewStyle().Bold(true).Foreground(SubtleColor)
)

// ---------------------------------------------------------------------------
// Glyph sets
// ---------------------------------------------------------------------------

type glyphSet struct {
	check    string
	cross    string
	bullet   string
	pending  string
	cursor   string
	markOn   string
	markOff  string
	barFull  string
	barEmpty string
	rule     string
	boxTL    string
	boxTR    string
	boxBL    string
	boxBR    string
	boxH     string
	boxV     string
	spinner  []string
}

func unicodeGlyphs() glyphSet {
	return glyphSet{
		check:    "✓",
		cross:    "✗",
		bullet:   "›",
		pending:  "○",
		cursor:   "❯",
		markOn:   "▣",
		markOff:  "▢",
		barFull:  "█",
		barEmpty: "░",
		rule:     "─",
		boxTL:    "╭",
		boxTR:    "╮",
		boxBL:    "╰",
		boxBR:    "╯",
		boxH:     "─",
		boxV:     "│",
		spinner:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	}
}

func asciiGlyphs() glyphSet {
	return glyphSet{
		check:    "+",
		cross:    "x",
		bullet:   ">",
		pending:  ".",
		cursor:   ">",
		markOn:   "[x]",
		markOff:  "[ ]",
		barFull:  "#",
		barEmpty: "-",
		rule:     "-",
		boxTL:    "+",
		boxTR:    "+",
		boxBL:    "+",
		boxBR:    "+",
		boxH:     "-",
		boxV:     "|",
		spinner:  []string{"|", "/", "-", "\\"},
	}
}

// ---------------------------------------------------------------------------
// Theme
// ---------------------------------------------------------------------------

// theme bundles the runtime presentation choices for the UI.
type theme struct {
	glyphs glyphSet
}

func newTheme(ascii bool) theme {
	g := unicodeGlyphs()
	if ascii {
		g = asciiGlyphs()
	}
	return theme{glyphs: g}
}

// stateColor maps a UI state to its accent color.
func stateColor(s state) color.Color {
	switch s {
	case stateUpdating:
		return PrimaryColor
	case stateErrorPrompt:
		return ErrorColor
	case stateConfirmCancel:
		return WarningColor
	case stateDone:
		return SuccessColor
	default:
		return SecondaryColor
	}
}

// ---------------------------------------------------------------------------
// Output helpers
//
// Lip Gloss v2 renders full-fidelity ANSI; downsampling to the terminal's
// actual color profile happens at print time, so commands outside the TUI must
// print through these helpers rather than fmt.
// ---------------------------------------------------------------------------

// Println writes to stdout with profile-aware color downsampling.
func Println(v ...any) {
	_, _ = lipgloss.Println(v...)
}

// Printf writes formatted output to stdout with color downsampling.
func Printf(format string, v ...any) {
	_, _ = lipgloss.Printf(format, v...)
}

// Fprintf writes formatted output to w with color downsampling.
func Fprintf(w io.Writer, format string, v ...any) {
	_, _ = lipgloss.Fprintf(w, format, v...)
}

// Fprintln writes to w with color downsampling.
func Fprintln(w io.Writer, v ...any) {
	_, _ = lipgloss.Fprintln(w, v...)
}

// ---------------------------------------------------------------------------
// Compat helpers used by the plain CLI output paths
// ---------------------------------------------------------------------------

// Brand returns the styled ASCII wordmark.
func Brand() string {
	return HighlightStyle.Render(logoArtText())
}

// MiniLogo returns a compact one-line brand.
func MiniLogo() string {
	return HighlightStyle.Render("⟨W⟩") + BoldStyle.Render(" winitrix")
}

// Divider returns a thin decorative line.
func Divider() string {
	return MutedStyle.Render("  ─────────────────────────────────────────────")
}

// Checkmark returns a styled success glyph.
func Checkmark() string {
	return SuccessStyle.Render("  ✓")
}

// Crossmark returns a styled failure glyph.
func Crossmark() string {
	return ErrorStyle.Render("  ✗")
}

// Bullet returns a styled accent bullet.
func Bullet() string {
	return AccentStyle.Render("  ›")
}

// Pending returns a styled pending glyph.
func Pending() string {
	return MutedStyle.Render("  ○")
}

func logoArtText() string {
	return strings.Join(logoArt, "\n")
}
