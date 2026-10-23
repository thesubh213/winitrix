//go:build windows

package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// processAlive reports whether a PID is currently running.
func processAlive(pid int) bool {
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

func waitProcessGone(t *testing.T, pid int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("process %d is still running after cancellation", pid)
}

func killIfAlive(pid int) {
	if processAlive(pid) {
		_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid)).Run()
	}
}

// parseMarker extracts the PID from a "PREFIX <pid>" output line.
func parseMarker(line, prefix string) (int, bool) {
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return 0, false
	}
	fields := strings.Fields(line[idx+len(prefix):])
	if len(fields) == 0 {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// TestCancellationKillsProcessTree is the regression test for "Ctrl+C leaves
// package managers running": cancelling the context must terminate the direct
// child AND its descendants.
func TestCancellationKillsProcessTree(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// PowerShell spawns a long-lived grandchild and reports its PID.
	const script = `$p = Start-Process -FilePath ping -ArgumentList '-n','120','127.0.0.1' -PassThru -WindowStyle Hidden; Write-Output ("GRANDCHILD " + $p.Id); Start-Sleep 120`

	lines := make(chan string, 64)
	done := make(chan StreamResult, 1)

	go func() {
		done <- RunWithProgress(ctx, func(line string) {
			select {
			case lines <- line:
			default:
			}
		}, "powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	}()

	var grandchild int
	timeout := time.After(30 * time.Second)
	for grandchild == 0 {
		select {
		case line := <-lines:
			if pid, ok := parseMarker(line, "GRANDCHILD "); ok {
				grandchild = pid
			}
		case <-timeout:
			cancel()
			<-done
			t.Fatal("timed out waiting for the grandchild process to start")
		}
	}
	defer killIfAlive(grandchild)

	if !processAlive(grandchild) {
		t.Fatalf("grandchild %d is not running before cancellation", grandchild)
	}

	start := time.Now()
	cancel()

	select {
	case res := <-done:
		if elapsed := time.Since(start); elapsed > 15*time.Second {
			t.Fatalf("cancellation took %s, expected it to be prompt", elapsed)
		}
		if res.Err == nil {
			t.Fatal("expected the run to report an error after cancellation")
		}
	case <-time.After(20 * time.Second):
		t.Fatal("RunWithProgress did not return after cancellation")
	}

	waitProcessGone(t, grandchild, 10*time.Second)
}
