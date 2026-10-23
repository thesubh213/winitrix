//go:build windows

package tui

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/runner"
)

func tuiProcessAlive(pid int) bool {
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH", "/FO", "CSV").Output()
	if err != nil {
		return false
	}
	want := fmt.Sprintf("\"%d\"", pid)
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}

func tuiKill(pid int) {
	if tuiProcessAlive(pid) {
		_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run()
	}
}

// procManager runs a real long-lived child process so cancellation can be
// verified end to end.
type procManager struct {
	name  string
	pidCh chan int
}

func (p *procManager) Name() string { return p.name }

func (p *procManager) Detect(ctx context.Context) bool { return true }

func (p *procManager) UpdateAll(ctx context.Context) ecosystem.Result {
	return p.UpdateAllWithProgress(ctx, nil)
}

func (p *procManager) UpdateAllWithProgress(ctx context.Context, cb runner.ProgressCallback) ecosystem.Result {
	wrapped := func(line string) {
		if pid, ok := parsePIDLine(line); ok {
			select {
			case p.pidCh <- pid:
			default:
			}
		}
		if cb != nil {
			cb(line)
		}
	}

	res := runner.RunWithProgress(ctx, wrapped, "powershell", "-NoProfile", "-NonInteractive",
		"-Command", `Write-Output ("CHILD " + $PID); Start-Sleep 120`)

	return ecosystem.Result{Success: res.Err == nil, Message: "finished", Error: res.Err}
}

func (p *procManager) Clean(ctx context.Context) ecosystem.Result {
	return ecosystem.Result{Success: true}
}
func (p *procManager) Doctor(ctx context.Context) ecosystem.Result {
	return ecosystem.Result{Success: true}
}

func parsePIDLine(line string) (int, bool) {
	idx := strings.Index(line, "CHILD ")
	if idx < 0 {
		return 0, false
	}
	fields := strings.Fields(line[idx+len("CHILD "):])
	if len(fields) == 0 {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// TestCancelConfirmedKillsRealProcess drives the real TUI cancel flow
// (ctrl+c -> confirm) against a live child process and asserts that the
// process is terminated.
func TestCancelConfirmedKillsRealProcess(t *testing.T) {
	mgr := &procManager{name: "ProcMgr", pidCh: make(chan int, 1)}

	opts := testOptions([]ecosystem.PackageManager{mgr})
	opts.TimeoutMinutes = 5

	m := NewModel(context.Background(), opts)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	var pid int
	select {
	case pid = <-mgr.pidCh:
	case <-time.After(30 * time.Second):
		_ = tm.Quit()
		t.Fatal("the manager never started a child process")
	}
	defer tuiKill(pid)

	if !tuiProcessAlive(pid) {
		t.Fatalf("child process %d is not running before cancellation", pid)
	}

	tm.Send(ctrlC())
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Cancel package updates"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(runeKey('y'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(20*time.Second))

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if !tuiProcessAlive(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("child process %d survived the confirmed cancellation", pid)
}
