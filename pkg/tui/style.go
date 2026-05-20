package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors — Curated deep-ocean palette
	PrimaryColor   = lipgloss.Color("#00BFFF") // Deep Sky Blue
	SecondaryColor = lipgloss.Color("#1E90FF") // Dodger Blue
	AccentColor    = lipgloss.Color("#5B9BD5") // Steel Blue
	SuccessColor   = lipgloss.Color("#00FF7F") // Spring Green
	ErrorColor     = lipgloss.Color("#FF4500") // Orange Red
	WarningColor   = lipgloss.Color("#FFD700") // Gold
	SubtleColor    = lipgloss.Color("#696969") // Dim Gray
	MutedColor     = lipgloss.Color("#4A4A4A") // Muted Gray
	BgColor        = lipgloss.Color("#0F0F0F")

	// Styles
	BaseStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Foreground(lipgloss.Color("#FFFFFF"))

	TitleStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 2).
			Align(lipgloss.Center).
			Bold(true).
			Foreground(PrimaryColor)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor).
			Padding(0, 2)

	SummaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 3)

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

// Brand returns the styled ASCII wordmark
func Brand() string {
	logo := `
 ██╗    ██╗██╗███╗   ██╗██╗████████╗██████╗ ██╗██╗  ██╗
 ██║    ██║██║████╗  ██║██║╚══██╔══╝██╔══██╗██║╚██╗██╔╝
 ██║ █╗ ██║██║██╔██╗ ██║██║   ██║   ██████╔╝██║ ╚███╔╝ 
 ██║███╗██║██║██║╚██╗██║██║   ██║   ██╔══██╗██║ ██╔██╗ 
 ╚███╔███╔╝██║██║ ╚████║██║   ██║   ██║  ██║██║██╔╝ ██╗
  ╚══╝╚══╝ ╚═╝╚═╝  ╚═══╝╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝╚═╝  ╚═╝`
	return HighlightStyle.Render(logo)
}

// MiniLogo returns a compact one-line brand
func MiniLogo() string {
	return HighlightStyle.Render("⟨W⟩") + BoldStyle.Render(" winitrix")
}

// Divider returns a thin decorative line
func Divider() string {
	return MutedStyle.Render("  ─────────────────────────────────────────────")
}

func Checkmark() string {
	return SuccessStyle.Render("  ✓")
}

func Crossmark() string {
	return ErrorStyle.Render("  ✗")
}

func Bullet() string {
	return AccentStyle.Render("  ›")
}

func Pending() string {
	return MutedStyle.Render("  ○")
}
