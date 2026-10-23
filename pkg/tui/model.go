package tui

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/logger"
	"github.com/thesubh213/winitrix/pkg/retry"
	"github.com/thesubh213/winitrix/pkg/runner"
)

type state int

const (
	stateScanning state = iota
	stateSelectManagers
	stateUpdating
	stateErrorPrompt
	stateConfirmCancel
	stateDone
)

// UpdateMsg contains the result of updating a single package manager.
type UpdateMsg struct {
	ManagerName string
	Result      ecosystem.Result
	Duration    time.Duration
}

// ProgressMsg contains progress information during package updates.
type ProgressMsg struct {
	ManagerName string
	PackageName string
	Percentage  int
	Status      string
}

// Options configures runtime behavior of the TUI model.
type Options struct {
	DryRun            bool
	TimeoutMinutes    int
	NonStop           bool
	Retries           int
	RetryDelay        time.Duration
	RetryBackoff      bool
	MaxRetryDelay     time.Duration
	Only              []string
	Exclude           []string
	SelectManagers    bool
	PreviousDurations map[string]time.Duration
	ProgressRelay     *ProgressRelay

	// Managers, when non-empty, skips ecosystem detection entirely and
	// updates the provided managers (used by --demo and tests).
	Managers []ecosystem.PackageManager

	Animations bool
	AltScreen  bool
	ASCII      bool
	Version    string
}

type scanDoneMsg struct {
	managers []ecosystem.PackageManager
}

// Model is the Bubble Tea model for the winitrix dashboard.
type Model struct {
	state state
	opts  Options
	ctx   context.Context

	// cancelAll cancels the model's root context (scan + running manager).
	cancelAll context.CancelFunc

	width  int
	height int

	startTime time.Time
	doneAt    time.Time
	theme     theme
	fx        *fx

	detectedManagers  []ecosystem.PackageManager
	results           []UpdateMsg
	previousDurations map[string]time.Duration
	expectedTotal     time.Duration

	currentManagerIdx  int
	currentPackageName string
	currentProgress    int
	currentStatus      string
	currentStart       time.Time
	cancelCtx          context.CancelFunc
	pendingFailure     *UpdateMsg
	holdNext           bool
	cancelled          bool
	note               string

	selectIdx int
	selection []bool

	// Details pane (live output of the running manager).
	expanded   bool
	logManager string
	logLines   []string

	scanProbed int
	scanTotal  int
	scanFound  []string

	borderColors []color.Color
	pulseStep    int

	progressRelay *ProgressRelay
}

// NewModel creates the TUI model.
func NewModel(ctx context.Context, opts Options) Model {
	if opts.TimeoutMinutes <= 0 {
		opts.TimeoutMinutes = 10
	}

	t := newTheme(opts.ASCII)

	runCtx, cancelRun := context.WithCancel(ctx)

	return Model{
		state:             stateScanning,
		ctx:               runCtx,
		cancelAll:         cancelRun,
		opts:              opts,
		theme:             t,
		fx:                newFX(t, opts.Animations),
		startTime:         time.Now(),
		width:             80,
		height:            24,
		progressRelay:     opts.ProgressRelay,
		previousDurations: opts.PreviousDurations,
	}
}

// Init starts the animation loop and the initial scan (or injected managers).
func (m Model) Init() tea.Cmd {
	if len(m.opts.Managers) > 0 {
		managers := m.opts.Managers
		return tea.Batch(tickCmd(m.animating()), func() tea.Msg {
			return scanDoneMsg{managers: managers}
		})
	}
	return tea.Batch(tickCmd(m.animating()), m.scanManagers)
}

func (m Model) animating() bool {
	return m.opts.Animations && m.state != stateDone
}

func (m Model) scanManagers() tea.Msg {
	logger.Info("Starting ecosystem scan")
	var cb func(found string, probed, total int)
	if m.progressRelay != nil {
		cb = m.progressRelay.ScanCallback()
	}
	managers := ecosystem.DetectFiltered(m.ctx, m.opts.Only, m.opts.Exclude, cb)
	logger.Info("Detected %d ecosystems", len(managers))
	return scanDoneMsg{managers: managers}
}

// Update handles all messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = max(msg.Width, 20)
		m.height = max(msg.Height, 8)
		return m, nil

	case animTickMsg:
		m.fx.frame++
		if m.pulseStep < len(m.borderColors) {
			m.pulseStep++
		}
		if m.state == stateDone {
			// The summary screen is static; stop the clock.
			return m, nil
		}
		return m, tickCmd(m.animating())

	case scanDoneMsg:
		m.scanTotal = len(msg.managers)
		m.scanProbed = m.scanTotal
		m.detectedManagers = msg.managers
		if len(m.detectedManagers) == 0 {
			m.transition(stateDone)
			return m, nil
		}
		m.selection = make([]bool, len(m.detectedManagers))
		for i := range m.selection {
			m.selection[i] = true
		}
		m.computeExpected()
		if m.opts.SelectManagers {
			m.transition(stateSelectManagers)
			return m, nil
		}
		m.transition(stateUpdating)
		return m.startNext()

	case ScanProgressMsg:
		m.scanProbed = msg.Probed
		m.scanTotal = msg.Total
		if msg.Found != "" {
			m.scanFound = append(m.scanFound, msg.Found)
		}
		return m, nil

	case ProgressMsg:
		if m.currentManagerIdx < len(m.detectedManagers) {
			currentName := m.detectedManagers[m.currentManagerIdx].Name()
			if strings.EqualFold(currentName, msg.ManagerName) {
				m.currentPackageName = msg.PackageName
				m.currentProgress = msg.Percentage
				m.currentStatus = msg.Status
			}
		}
		return m, nil

	case LogMsg:
		// Ignore late snapshots from a manager that already finished.
		if m.expanded && strings.EqualFold(msg.ManagerName, m.logManager) {
			m.logLines = msg.Lines
		}
		return m, nil

	case UpdateMsg:
		return m.handleUpdateResult(msg)

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleUpdateResult(msg UpdateMsg) (tea.Model, tea.Cmd) {
	m.currentPackageName = ""
	m.currentProgress = 0
	m.currentStatus = ""

	if m.cancelled {
		// The program is shutting down after a confirmed cancel; late results
		// are dropped.
		return m, nil
	}

	m.results = append(m.results, msg)
	m.currentManagerIdx++

	if !msg.Result.Success && !m.opts.NonStop {
		failure := msg
		m.pendingFailure = &failure
		m.transition(stateErrorPrompt)
		if m.expanded {
			m.seedLog()
		}
		return m, nil
	}

	if m.currentManagerIdx >= len(m.detectedManagers) {
		m.transition(stateDone)
		return m, nil
	}

	if m.state == stateConfirmCancel {
		// Hold the next manager until the user answers the cancel dialog.
		m.holdNext = true
		return m, nil
	}

	m.transition(stateUpdating)
	return m.startNext()
}

func (m Model) startNext() (tea.Model, tea.Cmd) {
	if m.currentManagerIdx >= len(m.detectedManagers) {
		m.transition(stateDone)
		return m, nil
	}

	manager := m.detectedManagers[m.currentManagerIdx]
	managerCtx, cancel := context.WithCancel(m.ctx)
	m.cancelCtx = cancel
	m.currentStart = time.Now()

	// A fresh manager gets a fresh log buffer.
	if m.progressRelay != nil {
		m.progressRelay.ResetLog(manager.Name())
	}
	if m.expanded {
		m.logManager = manager.Name()
		m.logLines = nil
	}

	progressCb := m.progressCallback(manager.Name())
	dryRun := m.opts.DryRun
	showDelay := m.opts.Animations

	return m, func() tea.Msg {
		start := time.Now()
		logger.Info("Starting update for %s", manager.Name())

		if dryRun {
			if showDelay {
				select {
				case <-time.After(150 * time.Millisecond):
				case <-managerCtx.Done():
				}
			}
			cancel()
			return UpdateMsg{
				ManagerName: manager.Name(),
				Result:      ecosystem.Result{Success: true, Message: "skipped (dry-run)"},
				Duration:    time.Since(start),
			}
		}

		res := m.runUpdateWithRetries(managerCtx, manager, progressCb)
		cancel()
		return UpdateMsg{
			ManagerName: manager.Name(),
			Result:      res,
			Duration:    time.Since(start),
		}
	}
}

func (m Model) progressCallback(managerName string) runner.ProgressCallback {
	if m.progressRelay == nil {
		return nil
	}
	return m.progressRelay.Callback(managerName)
}

func (m Model) runUpdateWithRetries(ctx context.Context, manager ecosystem.PackageManager, progressCb runner.ProgressCallback) ecosystem.Result {
	attempts := m.opts.Retries + 1
	if attempts < 1 {
		attempts = 1
	}

	var res ecosystem.Result
	for attempt := 1; attempt <= attempts; attempt++ {
		timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(m.opts.TimeoutMinutes)*time.Minute)
		res = manager.UpdateAllWithProgress(timeoutCtx, progressCb)
		cancel()

		if res.Success || attempt == attempts || ctx.Err() != nil {
			if !res.Success {
				switch {
				case errors.Is(ctx.Err(), context.Canceled):
					res.Message = "cancelled"
				case attempt > 1:
					res.Message = fmt.Sprintf("%s (failed after %d attempts)", res.Message, attempt)
				}
			}
			return res
		}

		delay := retry.ComputeDelay(attempt+1, m.opts.RetryDelay, m.opts.RetryBackoff, m.opts.MaxRetryDelay)
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return res
			}
		}
	}

	return res
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.state {
	case stateScanning:
		switch key {
		case "ctrl+c", "q", "Q", "esc":
			logger.Info("User interrupted during scan")
			m.cancelled = true
			m.note = "Cancelled — no updates were run."
			if m.cancelAll != nil {
				m.cancelAll()
			}
			m.transition(stateDone)
			return m, nil
		}

	case stateSelectManagers:
		switch key {
		case "up", "k":
			if m.selectIdx > 0 {
				m.selectIdx--
			}
		case "down", "j":
			if m.selectIdx < len(m.selection)-1 {
				m.selectIdx++
			}
		case "home":
			m.selectIdx = 0
		case "end":
			if n := len(m.selection); n > 0 {
				m.selectIdx = n - 1
			}
		case "space":
			if len(m.selection) > 0 {
				m.selection[m.selectIdx] = !m.selection[m.selectIdx]
			}
		case "a", "A":
			for i := range m.selection {
				m.selection[i] = true
			}
		case "n", "N":
			for i := range m.selection {
				m.selection[i] = false
			}
		case "enter", "r", "R":
			selected := make([]ecosystem.PackageManager, 0, len(m.detectedManagers))
			for i, mgr := range m.detectedManagers {
				if i < len(m.selection) && m.selection[i] {
					selected = append(selected, mgr)
				}
			}
			m.detectedManagers = selected
			m.currentManagerIdx = 0
			if len(m.detectedManagers) == 0 {
				m.note = "No ecosystems selected."
				m.transition(stateDone)
				return m, nil
			}
			m.transition(stateUpdating)
			return m.startNext()
		case "q", "Q", "esc", "ctrl+c":
			logger.Info("User quit from selection")
			m.cancelled = true
			m.note = "Cancelled — no updates were run."
			if m.cancelAll != nil {
				m.cancelAll()
			}
			m.transition(stateDone)
			return m, nil
		}

	case stateUpdating:
		switch key {
		case "d", "D":
			m.toggleDetails()
			return m, nil
		case "ctrl+c", "q", "Q":
			logger.Info("User requested cancellation - awaiting confirmation")
			m.transition(stateConfirmCancel)
			return m, nil
		case "esc":
			if m.expanded {
				m.collapseDetails()
				return m, nil
			}
			logger.Info("User requested cancellation - awaiting confirmation")
			m.transition(stateConfirmCancel)
			return m, nil
		}

	case stateErrorPrompt:
		switch key {
		case "d", "D":
			m.toggleDetails()
			return m, nil
		case "r", "R", "enter":
			m.pendingFailure = nil
			m.collapseDetails()
			// Drop the recorded failure and re-run the same manager.
			if n := len(m.results); n > 0 && m.currentManagerIdx > 0 {
				m.results = m.results[:n-1]
				m.currentManagerIdx--
			}
			m.transition(stateUpdating)
			return m.startNext()
		case "s", "S":
			m.pendingFailure = nil
			m.collapseDetails()
			if m.currentManagerIdx >= len(m.detectedManagers) {
				m.transition(stateDone)
				return m, nil
			}
			m.transition(stateUpdating)
			return m.startNext()
		case "esc":
			if m.expanded {
				m.collapseDetails()
				return m, nil
			}
			logger.Info("User quit after error")
			m.pendingFailure = nil
			m.transition(stateDone)
			return m, nil
		case "q", "Q", "ctrl+c":
			logger.Info("User quit after error")
			m.pendingFailure = nil
			m.transition(stateDone)
			return m, nil
		}

	case stateConfirmCancel:
		switch key {
		case "y", "Y", "enter", "r", "R":
			logger.Info("User confirmed cancellation")
			m.cancelUpdates()
			m.transition(stateDone)
			return m, tea.Quit
		case "n", "N", "esc", "s", "S", "q", "Q":
			logger.Info("User declined cancellation - resuming")
			if m.holdNext {
				m.holdNext = false
				m.transition(stateUpdating)
				return m.startNext()
			}
			m.transition(stateUpdating)
			return m, nil
		}

	case stateDone:
		return m, tea.Quit
	}

	return m, nil
}

// activeManagerName returns the manager whose output the details pane shows.
func (m Model) activeManagerName() string {
	if m.state == stateErrorPrompt && m.pendingFailure != nil {
		return m.pendingFailure.ManagerName
	}
	if m.currentManagerIdx < len(m.detectedManagers) {
		return m.detectedManagers[m.currentManagerIdx].Name()
	}
	return ""
}

// toggleDetails expands or collapses the live output pane.
func (m *Model) toggleDetails() {
	if m.expanded {
		m.collapseDetails()
		return
	}
	m.expanded = true
	m.seedLog()
	if m.progressRelay != nil {
		m.progressRelay.SubscribeLogs(true)
	}
}

func (m *Model) collapseDetails() {
	m.expanded = false
	m.logManager = ""
	m.logLines = nil
	if m.progressRelay != nil {
		m.progressRelay.SubscribeLogs(false)
	}
}

// seedLog loads the current tail of the active manager's retained output.
func (m *Model) seedLog() {
	name := m.activeManagerName()
	m.logManager = name
	m.logLines = nil
	if name == "" || m.progressRelay == nil {
		return
	}
	m.logLines = m.progressRelay.LogTail(name)
}

func (m *Model) cancelUpdates() {
	m.cancelled = true
	if m.cancelCtx != nil {
		m.cancelCtx()
	}
	if m.cancelAll != nil {
		m.cancelAll()
	}
	// Record the interrupted manager so the report reflects the partial run.
	if !m.holdNext && m.currentManagerIdx < len(m.detectedManagers) {
		name := m.detectedManagers[m.currentManagerIdx].Name()
		m.results = append(m.results, UpdateMsg{
			ManagerName: name,
			Result:      ecosystem.Result{Success: false, Message: "cancelled by user"},
			Duration:    time.Since(m.currentStart),
		})
		m.currentManagerIdx++
	}
}

// transition moves to a new state, arming a short border pulse.
func (m *Model) transition(s state) {
	if s == m.state {
		return
	}
	from := m.borderColor()
	m.state = s
	if s == stateDone && m.doneAt.IsZero() {
		m.doneAt = time.Now()
	}
	if m.opts.Animations {
		m.borderColors = pulseColors(from, stateColor(s), pulseSteps)
		m.pulseStep = 0
	}
}

func (m Model) borderColor() color.Color {
	if m.pulseStep < len(m.borderColors) {
		return m.borderColors[m.pulseStep]
	}
	return stateColor(m.state)
}

func (m *Model) computeExpected() {
	m.expectedTotal = 0
	for _, mgr := range m.detectedManagers {
		if d, ok := m.previousDurations[mgr.Name()]; ok {
			m.expectedTotal += d
		}
	}
}

func (m Model) estimatedRemaining() (time.Duration, bool) {
	if m.expectedTotal == 0 {
		return 0, false
	}
	remaining := m.expectedTotal - time.Since(m.startTime)
	if remaining <= 0 {
		return 0, false
	}
	return remaining, true
}

func (m Model) overallPercent() (int, bool) {
	total := len(m.detectedManagers)
	if total == 0 {
		return 0, false
	}
	frac := 0.0
	if m.currentManagerIdx < total && m.currentProgress > 0 {
		frac = float64(m.currentProgress) / 100
	}
	pct := int((float64(m.currentManagerIdx)+frac)/float64(total)*100 + 0.5)
	return min(max(pct, 0), 100), true
}

// GetResults returns the execution results to be printed cleanly on exit.
func (m Model) GetResults() []UpdateMsg {
	return m.results
}
