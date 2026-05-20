package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

// ProgressMsg contains progress information during package updates
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
}

type scanDoneMsg struct {
	managers []ecosystem.PackageManager
}

type Model struct {
	spinner   spinner.Model
	state     state
	opts      Options
	startTime time.Time

	detectedManagers  []ecosystem.PackageManager
	results           []UpdateMsg
	previousDurations map[string]time.Duration

	currentManagerIdx  int
	currentPackageName string
	currentProgress    int
	currentStatus      string
	currentStart       time.Time
	ctx                context.Context
	cancelCtx          context.CancelFunc
	progressRelay      *ProgressRelay
	pendingFailure     *UpdateMsg
	selectIdx          int
	selection          []bool
}

func NewModel(ctx context.Context, opts Options) Model {
	if opts.TimeoutMinutes <= 0 {
		opts.TimeoutMinutes = 10
	}

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(PrimaryColor)

	return Model{
		spinner:           s,
		state:             stateScanning,
		ctx:               ctx,
		opts:              opts,
		startTime:         time.Now(),
		progressRelay:     opts.ProgressRelay,
		previousDurations: opts.PreviousDurations,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.scanManagers,
	)
}

func (m Model) scanManagers() tea.Msg {
	logger.Info("Starting ecosystem scan")
	time.Sleep(800 * time.Millisecond) // Brief pause for visual polish
	managers := ecosystem.GetDetectedManagers(m.ctx)
	managers = ecosystem.FilterManagers(managers, m.opts.Only, m.opts.Exclude)
	logger.Info("Detected %d ecosystems", len(managers))
	return scanDoneMsg{managers: managers}
}

func (m Model) runNextUpdate() tea.Cmd {
	if m.currentManagerIdx >= len(m.detectedManagers) {
		return func() tea.Msg {
			return tea.Quit()
		}
	}

	manager := m.detectedManagers[m.currentManagerIdx]
	return func() tea.Msg {
		start := time.Now()
		logger.Info("Starting update for %s", manager.Name())

		// Handle dry-run mode
		if m.opts.DryRun {
			logger.Info("Dry-run: skipping %s", manager.Name())
			time.Sleep(200 * time.Millisecond) // Small visual delay
			return UpdateMsg{
				ManagerName: manager.Name(),
				Result: ecosystem.Result{
					Success: true,
					Message: "skipped (dry-run)",
					Updated: 0,
				},
				Duration: time.Since(start),
			}
		}

		// Use configurable timeout
		res := m.runUpdateWithRetries(manager, m.progressCallback(manager.Name()))
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

func (m Model) runUpdateWithRetries(manager ecosystem.PackageManager, progressCb runner.ProgressCallback) ecosystem.Result {
	attempts := m.opts.Retries + 1
	if attempts < 1 {
		attempts = 1
	}

	var res ecosystem.Result
	for attempt := 1; attempt <= attempts; attempt++ {
		timeoutCtx, cancel := context.WithTimeout(m.ctx, time.Duration(m.opts.TimeoutMinutes)*time.Minute)
		res = manager.UpdateAllWithProgress(timeoutCtx, progressCb)
		cancel()

		if res.Success || attempt == attempts {
			if !res.Success && attempt > 1 {
				res.Message = fmt.Sprintf("%s (failed after %d attempts)", res.Message, attempt)
			}
			return res
		}

		baseDelay := m.opts.RetryDelay
		delay := retry.ComputeDelay(attempt+1, baseDelay, m.opts.RetryBackoff, m.opts.MaxRetryDelay)
		if delay > 0 {
			time.Sleep(delay)
		}
	}

	return res
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case scanDoneMsg:
		m.detectedManagers = msg.managers
		if len(m.detectedManagers) == 0 {
			m.state = stateDone
			return m, tea.Quit
		}
		m.selection = make([]bool, len(m.detectedManagers))
		for i := range m.selection {
			m.selection[i] = true
		}
		if m.opts.SelectManagers {
			m.state = stateSelectManagers
			return m, nil
		}
		m.state = stateUpdating
		m.currentStart = time.Now()
		return m, m.runNextUpdate()

	case UpdateMsg:
		m.currentPackageName = ""
		m.currentProgress = 0
		m.currentStatus = ""

		if !msg.Result.Success {
			if m.opts.NonStop {
				m.results = append(m.results, msg)
				m.currentManagerIdx++
				if m.currentManagerIdx >= len(m.detectedManagers) {
					m.state = stateDone
					return m, tea.Quit
				}
				m.currentStart = time.Now()
				return m, m.runNextUpdate()
			}

			m.pendingFailure = &msg
			m.state = stateErrorPrompt
			return m, nil
		}

		m.results = append(m.results, msg)

		m.currentManagerIdx++

		if m.currentManagerIdx >= len(m.detectedManagers) {
			m.state = stateDone
			return m, tea.Quit
		}
		m.currentStart = time.Now()
		return m, m.runNextUpdate()

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

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.state == stateSelectManagers {
				if m.selectIdx > 0 {
					m.selectIdx--
				}
				return m, nil
			}
		case "down", "j":
			if m.state == stateSelectManagers {
				if m.selectIdx < len(m.selection)-1 {
					m.selectIdx++
				}
				return m, nil
			}
		case " ":
			if m.state == stateSelectManagers {
				if len(m.selection) > 0 {
					m.selection[m.selectIdx] = !m.selection[m.selectIdx]
				}
				return m, nil
			}
		case "a", "A":
			if m.state == stateSelectManagers {
				for i := range m.selection {
					m.selection[i] = true
				}
				return m, nil
			}
		case "n", "N":
			if m.state == stateSelectManagers {
				for i := range m.selection {
					m.selection[i] = false
				}
				return m, nil
			}
		case "ctrl+c":
			if m.state == stateUpdating {
				logger.Info("User pressed Ctrl+C - requesting cancellation confirmation")
				m.state = stateConfirmCancel
				return m, nil
			} else if m.state == stateConfirmCancel {
				// Already in confirm state, ignore second Ctrl+C
				return m, nil
			} else {
				logger.Info("User interrupted")
				return m, tea.Quit
			}
		case "q", "Q":
			if m.state == stateSelectManagers {
				logger.Info("User quit from selection")
				return m, tea.Quit
			}
			if m.state == stateErrorPrompt {
				logger.Info("User quit after error")
				if m.pendingFailure != nil {
					m.results = append(m.results, *m.pendingFailure)
					m.pendingFailure = nil
				}
				m.state = stateDone
				return m, tea.Quit
			}
			if m.state != stateConfirmCancel {
				logger.Info("User quit")
				return m, tea.Quit
			}
		case "r", "R", "enter":
			if m.state == stateSelectManagers {
				selected := make([]ecosystem.PackageManager, 0, len(m.detectedManagers))
				for i, mgr := range m.detectedManagers {
					if i < len(m.selection) && m.selection[i] {
						selected = append(selected, mgr)
					}
				}
				m.detectedManagers = selected
				m.currentManagerIdx = 0
				if len(m.detectedManagers) == 0 {
					m.state = stateDone
					return m, tea.Quit
				}
				m.state = stateUpdating
				m.currentStart = time.Now()
				return m, m.runNextUpdate()
			}
			if m.state == stateErrorPrompt {
				m.pendingFailure = nil
				m.state = stateUpdating
				m.currentStart = time.Now()
				return m, m.runNextUpdate()
			} else if m.state == stateConfirmCancel {
				logger.Info("User confirmed cancellation")
				if m.cancelCtx != nil {
					m.cancelCtx()
				}
				m.state = stateDone
				return m, tea.Quit
			}
		case "s", "S":
			if m.state == stateErrorPrompt {
				if m.pendingFailure != nil {
					m.results = append(m.results, *m.pendingFailure)
					m.pendingFailure = nil
				}
				m.currentManagerIdx++
				if m.currentManagerIdx >= len(m.detectedManagers) {
					m.state = stateDone
					return m, tea.Quit
				}
				m.state = stateUpdating
				m.currentStart = time.Now()
				return m, m.runNextUpdate()
			} else if m.state == stateConfirmCancel {
				logger.Info("User declined cancellation - resuming")
				m.state = stateUpdating
				return m, nil
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString(Brand())
	b.WriteString("\n")
	b.WriteString(SubtleStyle.Render("  Unified Windows Package Orchestration Engine"))
	b.WriteString("\n")
	b.WriteString(Divider())
	b.WriteString("\n\n")

	switch m.state {
	case stateScanning:
		b.WriteString(fmt.Sprintf("  %s Scanning installed ecosystems...\n", m.spinner.View()))
		b.WriteString(SubtleStyle.Render("    Detecting package managers on your system\n"))

	case stateSelectManagers:
		b.WriteString(DimBoldStyle.Render("  Select managers to update"))
		b.WriteString("\n\n")
		for i, mgr := range m.detectedManagers {
			cursor := " "
			if i == m.selectIdx {
				cursor = ">"
			}
			mark := "[ ]"
			if i < len(m.selection) && m.selection[i] {
				mark = "[x]"
			}
			line := fmt.Sprintf("  %s %s %s", cursor, mark, mgr.Name())
			if i == m.selectIdx {
				line = HighlightStyle.Render(line)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("\n")
		b.WriteString(SubtleStyle.Render("  [space] toggle  [a] all  [n] none  [enter] start  [q] quit"))

	case stateUpdating, stateErrorPrompt:
		// Progress indicator
		completed := m.currentManagerIdx
		total := len(m.detectedManagers)
		if m.state == stateErrorPrompt {
			completed = m.currentManagerIdx + 1
		}
		progress := fmt.Sprintf("[%d/%d]", completed, total)
		b.WriteString(fmt.Sprintf("  %s %s\n\n", DimBoldStyle.Render(progress), SubtleStyle.Render("Updating ecosystems")))
		elapsed := time.Since(m.startTime).Round(time.Second)
		if remaining, ok := m.estimatedRemaining(); ok {
			b.WriteString(SubtleStyle.Render(fmt.Sprintf("  Elapsed: %s  ETA: %s", elapsed, remaining.Round(time.Second))))
			b.WriteString("\n\n")
		} else {
			b.WriteString(SubtleStyle.Render(fmt.Sprintf("  Elapsed: %s", elapsed)))
			b.WriteString("\n\n")
		}

		for i, mgr := range m.detectedManagers {
			if i < m.currentManagerIdx {
				// Completed
				res := m.results[i]
				if res.Result.Success {
					if res.Result.Updated > 0 {
						b.WriteString(fmt.Sprintf("%s %s  %s\n",
							Checkmark(),
							BoldStyle.Render(mgr.Name()),
							SubtleStyle.Render(fmt.Sprintf("(%d updated)", res.Result.Updated))))
					} else {
						b.WriteString(fmt.Sprintf("%s %s  %s\n",
							Checkmark(),
							BoldStyle.Render(mgr.Name()),
							SubtleStyle.Render(res.Result.Message)))
					}
				} else {
					b.WriteString(fmt.Sprintf("%s %s  %s\n",
						Crossmark(),
						BoldStyle.Render(mgr.Name()),
						ErrorStyle.Render(res.Result.Message)))
				}
			} else if i == m.currentManagerIdx {
				if m.state == stateErrorPrompt {
					res := m.pendingFailure
					if res == nil {
						break
					}
					b.WriteString(fmt.Sprintf("%s %s  %s\n",
						Crossmark(),
						BoldStyle.Render(mgr.Name()),
						ErrorStyle.Render(res.Result.Message)))
					if res.Result.Error != nil {
						b.WriteString(fmt.Sprintf("    %s\n", SubtleStyle.Render(res.Result.Error.Error())))
					}
					b.WriteString("\n")
					b.WriteString(WarningStyle.Render(fmt.Sprintf("  ⚠ %s failed. ", mgr.Name())))
					b.WriteString(SubtleStyle.Render("[r]etry [s]kip [q]uit "))
				} else {
					// Currently updating
					b.WriteString(fmt.Sprintf("  %s %s\n",
						m.spinner.View(),
						HighlightStyle.Render(mgr.Name())))

					// Show current status and progress if available
					statusLine := strings.TrimSpace(m.currentStatus)
					if m.currentPackageName != "" {
						if statusLine != "" {
							statusLine = fmt.Sprintf("%s: %s", statusLine, m.currentPackageName)
						} else {
							statusLine = m.currentPackageName
						}
					}
					if statusLine != "" {
						if m.currentProgress > 0 {
							statusLine = fmt.Sprintf("%s (%d%%)", statusLine, m.currentProgress)
						}
						b.WriteString(fmt.Sprintf("    %s\n", SubtleStyle.Render(statusLine)))
					}
					elapsedCurrent := time.Since(m.currentStart).Round(time.Second)
					if expected := m.expectedDuration(mgr.Name()); expected > 0 {
						b.WriteString(fmt.Sprintf("    %s\n", SubtleStyle.Render(fmt.Sprintf("elapsed %s / ~%s", elapsedCurrent, expected.Round(time.Second)))))
					} else if !m.currentStart.IsZero() {
						b.WriteString(fmt.Sprintf("    %s\n", SubtleStyle.Render(fmt.Sprintf("elapsed %s", elapsedCurrent))))
					}
				}
			} else {
				// Pending
				b.WriteString(fmt.Sprintf("%s %s\n",
					Pending(),
					SubtleStyle.Render(mgr.Name())))
			}
		}

	case stateConfirmCancel:
		b.WriteString(fmt.Sprintf("  %s\n\n", WarningStyle.Render("⚠ Cancel package updates?")))
		b.WriteString(WarningStyle.Render("  This will stop the current operation.\n"))
		b.WriteString("\n")
		b.WriteString(SubtleStyle.Render("  Are you sure? "))
		b.WriteString(WarningStyle.Render("[y/N] "))

	case stateDone:
		if len(m.detectedManagers) == 0 {
			b.WriteString(WarningStyle.Render("  ⚠ No package managers detected on this system.\n"))
		} else {
			elapsed := time.Since(m.startTime).Round(time.Millisecond)
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				SuccessStyle.Render("Done"),
				SubtleStyle.Render(fmt.Sprintf("completed in %s", elapsed))))
		}
	}

	b.WriteString("\n")
	return b.String()
}

func (m Model) expectedDuration(managerName string) time.Duration {
	if m.previousDurations == nil {
		return 0
	}
	if d, ok := m.previousDurations[managerName]; ok {
		return d
	}
	return 0
}

func (m Model) estimatedRemaining() (time.Duration, bool) {
	if len(m.detectedManagers) == 0 || m.previousDurations == nil {
		return 0, false
	}

	var expectedTotal time.Duration
	for _, mgr := range m.detectedManagers {
		if d, ok := m.previousDurations[mgr.Name()]; ok {
			expectedTotal += d
		}
	}
	if expectedTotal == 0 {
		return 0, false
	}

	elapsed := time.Since(m.startTime)
	remaining := expectedTotal - elapsed
	if remaining <= 0 {
		return 0, false
	}

	return remaining, true
}

// GetResults returns the execution results to be printed cleanly on exit
func (m Model) GetResults() []UpdateMsg {
	return m.results
}
