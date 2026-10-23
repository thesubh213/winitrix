package tui

import (
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/runner"
)

// Log retention tuning for the expandable details pane.
const (
	logRingSize = 200                    // raw lines retained per manager
	logLineCap  = 512                    // bytes retained per line
	logThrottle = 120 * time.Millisecond // snapshot rate while subscribed
)

// ScanProgressMsg reports detection progress while scanning ecosystems.
type ScanProgressMsg struct {
	Found  string
	Probed int
	Total  int
}

// LogMsg carries a snapshot of retained raw output for a manager. Snapshots
// are only emitted while the details pane is subscribed.
type LogMsg struct {
	ManagerName string
	Lines       []string
}

// ProgressRelay bridges raw CLI output into structured TUI progress messages.
//
// Throttling happens *before* parsing: package managers emit thousands of
// spinner refresh lines, and running the parser on every one of them burns
// CPU for nothing. A stall guard guarantees the UI still refreshes at least
// every stallLimit even under a tight throttle window.
//
// Raw lines are always kept in a small bounded ring per manager so the
// details pane can be expanded at any time; full snapshots are only sent to
// the UI while the pane is open.
type ProgressRelay struct {
	mu          sync.Mutex
	send        func(tea.Msg)
	closed      bool
	lastParse   time.Time
	lastUpdate  time.Time
	lastManager string
	throttle    time.Duration
	stallLimit  time.Duration
	lastScan    time.Time

	logs    map[string][]string
	logSub  bool
	lastLog time.Time
}

// NewProgressRelay creates a relay with conservative throttling to avoid UI spam.
func NewProgressRelay() *ProgressRelay {
	return &ProgressRelay{
		throttle:   90 * time.Millisecond,
		stallLimit: 700 * time.Millisecond,
		logs:       make(map[string][]string),
	}
}

// Bind attaches the Bubble Tea send function.
func (p *ProgressRelay) Bind(send func(tea.Msg)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.send = send
}

// Close marks the relay closed so late worker goroutines drop their messages.
func (p *ProgressRelay) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	p.send = nil
}

// Callback returns a runner progress callback that emits structured progress messages.
func (p *ProgressRelay) Callback(managerName string) runner.ProgressCallback {
	return func(line string) {
		// Raw output is retained for the expandable details pane.
		p.recordLog(managerName, line)

		if len(line) < 4 {
			return
		}
		if !p.allow(managerName) {
			return
		}

		update, ok := ecosystem.ParseProgressLine(managerName, line)
		if !ok {
			return
		}

		p.emit(ProgressMsg{
			ManagerName: managerName,
			PackageName: update.PackageName,
			Percentage:  update.Percentage,
			Status:      update.Status,
		})
	}
}

// recordLog appends a raw line to the manager's bounded ring buffer and, while
// the details pane is subscribed, emits throttled snapshots.
func (p *ProgressRelay) recordLog(managerName, line string) {
	line = strings.TrimRight(line, " \t\r")
	if line == "" {
		return
	}
	if len(line) > logLineCap {
		line = line[:logLineCap]
	}

	p.mu.Lock()
	ring := p.logs[managerName]
	if n := len(ring); n == 0 || ring[n-1] != line {
		ring = append(ring, line)
		if len(ring) > logRingSize {
			ring = ring[len(ring)-logRingSize:]
		}
		p.logs[managerName] = ring
	}

	if !p.logSub || p.closed || p.send == nil {
		p.mu.Unlock()
		return
	}

	now := time.Now()
	if !p.lastLog.IsZero() && now.Sub(p.lastLog) < logThrottle {
		p.mu.Unlock()
		return
	}
	p.lastLog = now

	snapshot := make([]string, len(ring))
	copy(snapshot, ring)
	send := p.send
	p.mu.Unlock()

	send(LogMsg{ManagerName: managerName, Lines: snapshot})
}

// SubscribeLogs enables or disables LogMsg snapshots (details pane open/closed).
func (p *ProgressRelay) SubscribeLogs(on bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logSub = on
	p.lastLog = time.Time{}
}

// LogTail returns a copy of the retained raw output for a manager.
func (p *ProgressRelay) LogTail(managerName string) []string {
	p.mu.Lock()
	defer p.mu.Unlock()

	ring := p.logs[managerName]
	out := make([]string, len(ring))
	copy(out, ring)
	return out
}

// ResetLog clears retained output for a manager when its update starts.
func (p *ProgressRelay) ResetLog(managerName string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.logs, managerName)
}

// ScanCallback returns a detection progress callback for the scanning screen.
func (p *ProgressRelay) ScanCallback() func(found string, probed, total int) {
	return func(found string, probed, total int) {
		p.mu.Lock()
		send := p.send
		if p.closed || send == nil {
			p.mu.Unlock()
			return
		}
		now := time.Now()
		if probed < total && now.Sub(p.lastScan) < p.throttle {
			p.mu.Unlock()
			return
		}
		p.lastScan = now
		p.mu.Unlock()

		send(ScanProgressMsg{Found: found, Probed: probed, Total: total})
	}
}

// allow reports whether a line for the given manager should be parsed.
func (p *ProgressRelay) allow(managerName string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	if managerName != p.lastManager {
		p.lastManager = managerName
		p.lastParse = now
		p.lastUpdate = now
		return true
	}
	if now.Sub(p.lastParse) < p.throttle && now.Sub(p.lastUpdate) < p.stallLimit {
		return false
	}
	p.lastParse = now
	return true
}

func (p *ProgressRelay) emit(msg tea.Msg) {
	p.mu.Lock()
	if p.closed || p.send == nil {
		p.mu.Unlock()
		return
	}
	send := p.send
	p.lastUpdate = time.Now()
	p.mu.Unlock()

	send(msg)
}
