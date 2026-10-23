package tui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// truncate cuts a styled string to the given display width without breaking
// ANSI sequences or wide characters.
func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return ansi.Truncate(s, w, "")
}

// padRight pads a styled string with spaces to the given display width.
func padRight(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw >= w {
		return s
	}
	return s + strings.Repeat(" ", w-sw)
}

// padLeft pads a styled string with leading spaces to the given width.
func padLeft(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw >= w {
		return s
	}
	return strings.Repeat(" ", w-sw) + s
}

// windowRange returns a [start,end) window over total items that keeps cursor
// visible, for scrollable lists.
func windowRange(total, cursor, maxRows int) (int, int) {
	if maxRows < 1 || maxRows >= total {
		return 0, total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	start := cursor - maxRows/2
	if start < 0 {
		start = 0
	}
	if start+maxRows > total {
		start = total - maxRows
	}
	return start, start + maxRows
}

// renderPanel draws a rounded-border panel of exactly the given total width.
//
// The frame is rendered by hand rather than delegated to lipgloss borders:
// it is faster, keeps borders perfectly aligned, and sidesteps a lipgloss v2
// defect where boolean style properties above bit 31 are truncated on 32-bit
// builds (`attrs int`), which silently drops the left border on 386.
func renderPanel(rows []string, width int, borderColor color.Color, g glyphSet) string {
	if width < 6 {
		width = 6
	}
	inner := width - 4
	if inner < 1 {
		inner = 1
	}

	border := lipgloss.NewStyle().Foreground(borderColor)
	horizontal := strings.Repeat(g.boxH, inner+2)

	var b strings.Builder
	b.WriteString(border.Render(g.boxTL + horizontal + g.boxTR))
	b.WriteString("\n")
	for _, row := range rows {
		b.WriteString(border.Render(g.boxV))
		b.WriteString(" ")
		b.WriteString(padRight(truncate(row, inner), inner))
		b.WriteString(" ")
		b.WriteString(border.Render(g.boxV))
		b.WriteString("\n")
	}
	b.WriteString(border.Render(g.boxBL + horizontal + g.boxBR))

	return b.String()
}

// innerWidth is the usable content width inside a panel of the given width.
func innerWidth(panelWidth int) int {
	inner := panelWidth - 4
	if inner < 2 {
		inner = 2
	}
	return inner
}

// keyHints renders "  [key] label" pairs. Arguments alternate key/label.
func keyHints(pairs ...string) string {
	var b strings.Builder
	for i := 0; i+1 < len(pairs); i += 2 {
		if i > 0 {
			b.WriteString("   ")
		}
		b.WriteString(DimBoldStyle.Render("[" + pairs[i] + "]"))
		b.WriteString(SubtleStyle.Render(" " + pairs[i+1]))
	}
	return b.String()
}

// highlightRow renders a full-width row with the selection background.
func highlightRow(content string, width int) string {
	inner := innerWidth(width)
	row := padRight(truncate(content, inner), inner)
	return lipgloss.NewStyle().
		Background(highlightBg).
		Render(row)
}

// statusGlyph returns the row status glyph for a manager entry.
func statusGlyph(g glyphSet, status rowStatus) string {
	switch status {
	case rowDone:
		return SuccessStyle.Render(g.check)
	case rowFailed:
		return ErrorStyle.Render(g.cross)
	case rowActive:
		return "" // filled with the animated spinner
	default:
		return MutedStyle.Render(g.pending)
	}
}

type rowStatus int

const (
	rowPending rowStatus = iota
	rowActive
	rowDone
	rowFailed
)
