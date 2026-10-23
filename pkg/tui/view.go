package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

const maxContentWidth = 120

// View renders the dashboard. Content is assembled from precomputed pieces;
// per-frame work is limited to a handful of string concatenations.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = m.opts.AltScreen
	v.WindowTitle = "winitrix"

	if m.state == stateUpdating || m.state == stateErrorPrompt {
		if pct, ok := m.overallPercent(); ok {
			v.ProgressBar = tea.NewProgressBar(tea.ProgressBarDefault, pct)
		} else {
			v.ProgressBar = tea.NewProgressBar(tea.ProgressBarIndeterminate, 0)
		}
	}

	return v
}

func (m Model) contentWidth() int {
	w := m.width
	if w > maxContentWidth {
		w = maxContentWidth
	}
	if w < 20 {
		w = 20
	}
	return w
}

func (m Model) compactHeader() bool {
	return m.height < 26 || m.width < 62
}

func (m Model) render() string {
	width := m.contentWidth()
	var b strings.Builder

	b.WriteString(m.renderHeader(width))

	switch m.state {
	case stateScanning:
		m.renderScanning(&b, width)
	case stateSelectManagers:
		m.renderSelection(&b, width)
	case stateUpdating, stateErrorPrompt:
		m.renderUpdating(&b, width)
	case stateConfirmCancel:
		m.renderCancel(&b, width)
	case stateDone:
		m.renderDone(&b, width)
	}

	b.WriteString(m.renderFooter(width))
	return b.String()
}

// ---------------------------------------------------------------------------
// Header / footer
// ---------------------------------------------------------------------------

func (m Model) renderHeader(width int) string {
	var b strings.Builder
	b.WriteString("\n")

	if m.compactHeader() {
		b.WriteString(twoCols("  "+MiniLogo(), m.versionLabel(), width))
		b.WriteString("\n")
	} else {
		logo := m.fx.logoFrame()
		for _, line := range strings.Split(logo, "\n") {
			b.WriteString("  " + line + "\n")
		}
		left := "  " + SubtleStyle.Render("Unified Windows Package Orchestration Engine")
		b.WriteString(twoCols(left, m.versionLabel(), width))
		b.WriteString("\n")
	}

	b.WriteString("  " + m.fx.rule(width-2) + "\n\n")
	return b.String()
}

func (m Model) versionLabel() string {
	if m.opts.Version == "" {
		return ""
	}
	return SubtleStyle.Render("v" + m.opts.Version)
}

func (m Model) renderFooter(width int) string {
	var pairs []string
	switch m.state {
	case stateScanning:
		pairs = []string{"ctrl+c", "cancel"}
	case stateSelectManagers:
		pairs = []string{"space", "toggle", "a", "all", "n", "none", "enter", "start", "q", "quit"}
	case stateUpdating:
		if m.expanded {
			pairs = []string{"esc", "close", "ctrl+c", "cancel"}
		} else {
			pairs = []string{"d", "details", "ctrl+c", "cancel"}
		}
	case stateErrorPrompt:
		if m.expanded {
			pairs = []string{"esc", "close", "r", "retry", "s", "skip", "q", "quit"}
		} else {
			pairs = []string{"d", "details", "r", "retry", "s", "skip", "q", "quit"}
		}
	case stateConfirmCancel:
		pairs = []string{"y", "confirm", "n", "resume"}
	case stateDone:
		pairs = []string{"enter", "exit"}
	}

	// Add hint chips while they fit; never cut a chip in half.
	available := width - 2
	used := 0
	var b strings.Builder
	for i := 0; i+1 < len(pairs); i += 2 {
		chip := keyHints(pairs[i], pairs[i+1])
		chipW := ansi.StringWidth(chip)
		sep := 0
		if used > 0 {
			sep = 3
		}
		if used+sep+chipW > available {
			break
		}
		if sep > 0 {
			b.WriteString("   ")
		}
		b.WriteString(chip)
		used += sep + chipW
	}

	return "\n  " + b.String() + "\n"
}

// ---------------------------------------------------------------------------
// Screens
// ---------------------------------------------------------------------------

func (m Model) renderScanning(b *strings.Builder, width int) {
	right := ""
	if m.scanTotal > 0 {
		right = DimBoldStyle.Render(fmt.Sprintf("[%d/%d]", m.scanProbed, m.scanTotal))
	}
	left := "  " + m.fx.spinnerGlyph() + " " + HighlightStyle.Render("Scanning installed ecosystems")
	b.WriteString(twoCols(left, right, width))
	b.WriteString("\n\n")

	barW := width - 6
	if barW < 10 {
		barW = 10
	}
	b.WriteString("  " + m.fx.bar(0, barW, true))
	b.WriteString("\n")

	if len(m.scanFound) > 0 {
		found := strings.Join(m.scanFound, "  ")
		b.WriteString("  " + truncate(SubtleStyle.Render("found: "+found), width-4))
		b.WriteString("\n")
	}
}

func (m Model) renderSelection(b *strings.Builder, width int) {
	total := len(m.detectedManagers)
	right := SubtleStyle.Render(fmt.Sprintf("%d/%d selected", m.selectedCount(), total))
	b.WriteString(twoCols("  "+HighlightStyle.Render("Select ecosystems to update"), right, width))
	b.WriteString("\n\n")

	start, end := windowRange(total, m.selectIdx, m.listRows(2))

	rows := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		mgr := m.detectedManagers[i]
		mark := m.theme.glyphs.markOff
		selected := i < len(m.selection) && m.selection[i]
		if selected {
			mark = m.theme.glyphs.markOn
		}
		cursor := " "
		if i == m.selectIdx {
			cursor = m.theme.glyphs.cursor
		}

		if i == m.selectIdx {
			plain := fmt.Sprintf("%s %s %s", cursor, mark, mgr.Name())
			rows = append(rows, highlightRow(plain, width))
			continue
		}

		markStyled := MutedStyle.Render(mark)
		if selected {
			markStyled = SuccessStyle.Render(mark)
		}
		nameStyled := MutedStyle.Render(mgr.Name())
		if selected {
			nameStyled = BoldStyle.Render(mgr.Name())
		}
		rows = append(rows, fmt.Sprintf("%s %s %s", cursor, markStyled, nameStyled))
	}

	b.WriteString(renderPanel(rows, width, m.borderColor(), m.theme.glyphs))
	b.WriteString("\n")
}

func (m Model) renderUpdating(b *strings.Builder, width int) {
	total := len(m.detectedManagers)
	completed := m.currentManagerIdx

	title := "Updating ecosystems"
	if m.state == stateErrorPrompt {
		title = "Update paused on an error"
	}

	counter := DimBoldStyle.Render(fmt.Sprintf("[%d/%d]", min(completed, total), total))
	elapsed := formatDuration(time.Since(m.startTime))
	right := counter + "  "
	if remaining, ok := m.estimatedRemaining(); ok {
		right += SubtleStyle.Render(elapsed + " / ~" + formatDuration(remaining))
	} else {
		right += SubtleStyle.Render(elapsed)
	}
	b.WriteString(twoCols("  "+HighlightStyle.Render(title), right, width))
	b.WriteString("\n\n")

	inner := innerWidth(width)
	nameW := m.nameColumnWidth(inner)
	durW := 7
	detailW := inner - nameW - durW - 4
	if detailW < 10 {
		detailW = 10
	}

	extra := 2
	if m.expanded {
		extra += m.logPaneLines() + 4 // header + borders + spacing
	}

	start, end := windowRange(total, m.currentManagerIdx, m.listRows(extra))
	rows := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		rows = append(rows, m.updateRow(i, nameW, detailW, durW))
	}

	borderColor := m.borderColor()
	b.WriteString(renderPanel(rows, width, borderColor, m.theme.glyphs))
	b.WriteString("\n")

	if m.state == stateUpdating {
		pctTxt := ""
		frac := 0.0
		if pct, ok := m.overallPercent(); ok {
			frac = float64(pct) / 100
			pctTxt = strconv.Itoa(pct) + "%"
		}

		// Fit the bar and the label into the available width; drop the label
		// on very narrow terminals.
		label := "  " + DimBoldStyle.Render("Overall") + "  "
		reserved := ansi.StringWidth(label) + ansi.StringWidth(pctTxt) + 4
		barW := width - reserved
		if barW < 8 {
			label = "  "
			barW = width - 2 - ansi.StringWidth(pctTxt) - 2
		}
		if barW < 4 {
			barW = 4
		}

		line := label + m.fx.bar(frac, barW, false) + "  " + SubtleStyle.Render(pctTxt)
		b.WriteString(truncate(line, width))
		b.WriteString("\n")
	} else {
		m.renderErrorPrompt(b, width)
	}

	m.renderLogPane(b, width)
}

// logPaneLines is how many raw output lines the details pane shows.
func (m Model) logPaneLines() int {
	lines := m.height / 5
	if lines < 3 {
		lines = 3
	}
	if lines > 10 {
		lines = 10
	}
	return lines
}

// renderLogPane draws the expandable live output panel for the running manager.
func (m Model) renderLogPane(b *strings.Builder, width int) {
	if !m.expanded {
		return
	}

	name := m.activeManagerName()
	if name == "" {
		return
	}

	header := BoldStyle.Render(name)
	var meta []string
	if m.state == stateUpdating {
		if m.currentStatus != "" {
			meta = append(meta, m.currentStatus)
		}
		if m.currentPackageName != "" {
			meta = append(meta, m.currentPackageName)
		}
		if m.currentProgress > 0 {
			meta = append(meta, fmt.Sprintf("%d%%", m.currentProgress))
		}
		meta = append(meta, formatDuration(time.Since(m.currentStart)))
	} else if m.pendingFailure != nil {
		meta = append(meta, "failed")
	}
	if len(meta) > 0 {
		// Include metadata only while it fits, so narrow panes never show a
		// truncated fragment of a field.
		limit := innerWidth(width)
		used := ansi.StringWidth(header)
		var kept []string
		for _, item := range meta {
			itemW := ansi.StringWidth(item)
			sep := 0
			if len(kept) > 0 {
				sep = 5 // "  ·  "
			}
			if used+sep+itemW > limit {
				break
			}
			kept = append(kept, item)
			used += sep + itemW
		}
		if len(kept) > 0 {
			header += SubtleStyle.Render("  ·  " + strings.Join(kept, "  ·  "))
		}
	}

	visible := m.logPaneLines()
	lines := m.logLines
	if len(lines) > visible {
		lines = lines[len(lines)-visible:]
	}

	rows := make([]string, 0, visible+1)
	rows = append(rows, header)
	for _, line := range lines {
		rows = append(rows, styleLogLine(line))
	}
	// Keep the pane a fixed height so the layout doesn't jump.
	for len(rows) < visible+1 {
		rows = append(rows, "")
	}
	if len(lines) == 0 {
		rows[1] = MutedStyle.Render("waiting for output…")
	}

	b.WriteString("\n")
	b.WriteString(renderPanel(rows, width, m.borderColor(), m.theme.glyphs))
	b.WriteString("\n")
}

// styleLogLine renders one raw output line with light severity highlighting.
func styleLogLine(line string) string {
	clean := strings.TrimRight(ansi.Strip(line), " \t\r")
	if clean == "" {
		return ""
	}

	lower := strings.ToLower(clean)
	switch {
	case strings.Contains(lower, "error"), strings.Contains(lower, "failed"), strings.Contains(lower, "fatal"), strings.Contains(lower, "denied"):
		return ErrorStyle.Render(clean)
	case strings.Contains(lower, "warn"), strings.Contains(lower, "deprecated"):
		return WarningStyle.Render(clean)
	case strings.Contains(lower, "success"), strings.Contains(lower, "installed"), strings.Contains(lower, "updated"), strings.Contains(lower, "upgraded"):
		return SuccessStyle.Render(clean)
	default:
		return SubtleStyle.Render(clean)
	}
}

func (m Model) renderErrorPrompt(b *strings.Builder, width int) {
	fail := m.pendingFailure
	if fail == nil {
		return
	}

	b.WriteString("\n")
	b.WriteString("  " + WarningStyle.Render(truncate("⚠ "+fail.ManagerName+" failed", width-4)))
	b.WriteString("\n")
	b.WriteString("  " + truncate(ErrorStyle.Render(fail.Result.Message), width-4))
	b.WriteString("\n")
	if fail.Result.Error != nil {
		b.WriteString("  " + truncate(SubtleStyle.Render(fail.Result.Error.Error()), width-4))
		b.WriteString("\n")
	}
}

func (m Model) updateRow(i, nameW, detailW, durW int) string {
	mgr := m.detectedManagers[i]

	if i < len(m.results) {
		res := m.results[i]
		status := rowDone
		if !res.Result.Success {
			status = rowFailed
		}
		glyph := statusGlyph(m.theme.glyphs, status)

		var detail string
		if res.Result.Success {
			if res.Result.Updated > 0 {
				detail = SubtleStyle.Render(fmt.Sprintf("(%d updated)", res.Result.Updated))
			} else {
				detail = SubtleStyle.Render(res.Result.Message)
			}
		} else {
			detail = ErrorStyle.Render(res.Result.Message)
		}

		return fmt.Sprintf("%s %s %s %s",
			glyph,
			padRight(BoldStyle.Render(mgr.Name()), nameW),
			padRight(truncate(detail, detailW), detailW),
			padLeft(SubtleStyle.Render(formatDuration(res.Duration)), durW),
		)
	}

	if i == m.currentManagerIdx && m.state != stateErrorPrompt {
		return fmt.Sprintf("%s %s %s %s",
			m.fx.spinnerGlyph(),
			padRight(HighlightStyle.Render(mgr.Name()), nameW),
			padRight(m.activeDetail(detailW), detailW),
			padLeft(SubtleStyle.Render(formatDuration(time.Since(m.currentStart))), durW),
		)
	}

	return fmt.Sprintf("%s %s", statusGlyph(m.theme.glyphs, rowPending), MutedStyle.Render(mgr.Name()))
}

func (m Model) activeDetail(width int) string {
	barW := 18
	if width >= 44 {
		barW = 26
	}
	if barW > width-6 {
		barW = max(width-6, 6)
	}

	var parts []string
	if m.currentProgress > 0 {
		parts = append(parts, m.fx.bar(float64(m.currentProgress)/100, barW, false))
	}

	status := m.currentStatus
	if m.currentPackageName != "" {
		if status != "" {
			status = status + ": " + m.currentPackageName
		} else {
			status = m.currentPackageName
		}
	}
	if status == "" && m.currentProgress == 0 {
		status = "working…"
	}

	used := 0
	for _, p := range parts {
		used += ansi.StringWidth(p) + 1
	}
	if status != "" {
		remaining := width - used
		if m.currentProgress > 0 {
			remaining -= 5 // room for the percentage suffix
		}
		if remaining > 4 {
			parts = append(parts, SubtleStyle.Render(truncate(status, remaining)))
		}
	}
	if m.currentProgress > 0 {
		parts = append(parts, SubtleStyle.Render(fmt.Sprintf("%d%%", m.currentProgress)))
	}

	if len(parts) == 0 {
		return ""
	}
	return truncate(strings.Join(parts, " "), width)
}

func (m Model) renderCancel(b *strings.Builder, width int) {
	panelW := width - 2
	if panelW > 64 {
		panelW = 64
	}

	rows := []string{
		WarningStyle.Render("⚠  Cancel package updates?"),
		"",
		"Stopping now interrupts the current operation.",
		"Ecosystems that already finished keep their",
		"results in the final report.",
		"",
		keyHints("y", "yes, cancel", "n", "no, continue"),
	}

	b.WriteString(indentBlock(renderPanel(rows, panelW, WarningColor, m.theme.glyphs), "  "))
	b.WriteString("\n")
}

func (m Model) renderDone(b *strings.Builder, width int) {
	if len(m.detectedManagers) == 0 {
		msg := m.note
		if msg == "" {
			msg = "No package managers detected on this system."
		}
		b.WriteString("  " + WarningStyle.Render(truncate("⚠ "+msg, width-2)))
		b.WriteString("\n")
		return
	}

	if len(m.results) == 0 {
		msg := m.note
		if msg == "" {
			msg = "Cancelled — no updates were run."
		}
		b.WriteString("  " + WarningStyle.Render(truncate("⚠ "+msg, width-2)))
		b.WriteString("\n")
		return
	}

	success, failed, updated := m.totals()
	headline := "Done"
	if m.cancelled {
		headline = "Cancelled"
	}

	elapsed := time.Since(m.startTime)
	if !m.doneAt.IsZero() {
		elapsed = m.doneAt.Sub(m.startTime)
	}
	summary := fmt.Sprintf("%d ecosystems · %d updated · %d failed · %s",
		success+failed, updated, failed, formatDuration(elapsed))

	headlineStyle := SuccessStyle
	if failed > 0 || m.cancelled {
		headlineStyle = WarningStyle
	}
	b.WriteString(twoCols("  "+headlineStyle.Render("✓ "+headline), SubtleStyle.Render(summary), width))
	b.WriteString("\n\n")

	start, end := windowRange(len(m.results), len(m.results)-1, m.listRows(2))
	inner := innerWidth(width)
	nameW := m.nameColumnWidth(inner)
	detailW := inner - nameW - 7 - 4
	if detailW < 10 {
		detailW = 10
	}

	rows := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		res := m.results[i]
		glyph := statusGlyph(m.theme.glyphs, rowDone)
		var detail string
		if res.Result.Success {
			if res.Result.Updated > 0 {
				detail = SubtleStyle.Render(fmt.Sprintf("(%d updated)", res.Result.Updated))
			} else {
				detail = SubtleStyle.Render(res.Result.Message)
			}
		} else {
			glyph = statusGlyph(m.theme.glyphs, rowFailed)
			detail = ErrorStyle.Render(res.Result.Message)
		}
		rows = append(rows, fmt.Sprintf("%s %s %s %s",
			glyph,
			padRight(BoldStyle.Render(res.ManagerName), nameW),
			padRight(truncate(detail, detailW), detailW),
			padLeft(SubtleStyle.Render(formatDuration(res.Duration)), 7),
		))
	}

	borderColor := SuccessColor
	if failed > 0 || m.cancelled {
		borderColor = WarningColor
	}
	b.WriteString(renderPanel(rows, width, borderColor, m.theme.glyphs))
	b.WriteString("\n")
}

// ---------------------------------------------------------------------------
// Small helpers
// ---------------------------------------------------------------------------

func (m Model) selectedCount() int {
	n := 0
	for _, sel := range m.selection {
		if sel {
			n++
		}
	}
	return n
}

func (m Model) totals() (success, failed, updated int) {
	for _, res := range m.results {
		if res.Result.Success {
			success++
			updated += res.Result.Updated
		} else {
			failed++
		}
	}
	return success, failed, updated
}

// nameColumnWidth picks a stable name column width for the given inner width.
func (m Model) nameColumnWidth(inner int) int {
	longest := 0
	for _, mgr := range m.detectedManagers {
		if n := len(mgr.Name()); n > longest {
			longest = n
		}
	}
	w := longest + 1
	if w < 10 {
		w = 10
	}
	if maxW := inner - 24; w > maxW && maxW > 8 {
		w = maxW
	}
	return w
}

// listRows computes how many list rows fit in the current terminal height.
func (m Model) listRows(extra int) int {
	base := 19
	if m.compactHeader() {
		base = 14
	}
	rows := m.height - base - extra
	if rows < 3 {
		rows = 3
	}
	return rows
}

// twoCols joins a left and right column within the given width.
func twoCols(left, right string, width int) string {
	lw := ansi.StringWidth(left)
	rw := ansi.StringWidth(right)
	gap := width - lw - rw - 3
	if gap < 1 {
		return truncate(left, width)
	}
	return left + strings.Repeat(" ", gap) + right
}

// indentBlock prefixes every line of a block with the given indent.
func indentBlock(block, indent string) string {
	if indent == "" {
		return block
	}
	return indent + strings.ReplaceAll(block, "\n", "\n"+indent)
}
