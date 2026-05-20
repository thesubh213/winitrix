package tui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/runner"
)

// ProgressRelay bridges raw CLI output into structured TUI progress messages.
type ProgressRelay struct {
	mu       sync.Mutex
	send     func(tea.Msg)
	lastSent time.Time
	throttle time.Duration
}

// NewProgressRelay creates a relay with conservative throttling to avoid UI spam.
func NewProgressRelay() *ProgressRelay {
	return &ProgressRelay{throttle: 120 * time.Millisecond}
}

// Bind attaches the Bubble Tea send function.
func (p *ProgressRelay) Bind(send func(tea.Msg)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.send = send
}

// Callback returns a runner progress callback that emits structured progress messages.
func (p *ProgressRelay) Callback(managerName string) runner.ProgressCallback {
	return func(line string) {
		update, ok := ecosystem.ParseProgressLine(managerName, line)
		if !ok {
			return
		}

		msg := ProgressMsg{
			ManagerName: managerName,
			PackageName: update.PackageName,
			Percentage:  update.Percentage,
			Status:      update.Status,
		}

		p.sendMsg(msg)
	}
}

func (p *ProgressRelay) sendMsg(msg tea.Msg) {
	p.mu.Lock()
	send := p.send
	if send == nil {
		p.mu.Unlock()
		return
	}

	if p.throttle > 0 {
		if time.Since(p.lastSent) < p.throttle {
			p.mu.Unlock()
			return
		}
		p.lastSent = time.Now()
	}
	p.mu.Unlock()

	send(msg)
}
