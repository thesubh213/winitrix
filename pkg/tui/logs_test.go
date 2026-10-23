package tui

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
)

// collectLogMsgs binds a relay to a synchronised LogMsg collector.
func collectLogMsgs(relay *ProgressRelay) func() []LogMsg {
	var (
		mu   sync.Mutex
		msgs []LogMsg
	)
	relay.Bind(func(msg tea.Msg) {
		if lm, ok := msg.(LogMsg); ok {
			mu.Lock()
			msgs = append(msgs, lm)
			mu.Unlock()
		}
	})
	return func() []LogMsg {
		mu.Lock()
		defer mu.Unlock()
		out := make([]LogMsg, len(msgs))
		copy(out, msgs)
		return out
	}
}

func TestRelayRetainsLogLines(t *testing.T) {
	relay := NewProgressRelay()
	collect := collectLogMsgs(relay) // not subscribed: no messages expected
	cb := relay.Callback("Winget")

	cb("line one")
	cb("line two")
	cb("line two") // consecutive duplicate is dropped
	cb("")
	cb("   ")

	want := []string{"line one", "line two"}
	if got := relay.LogTail("Winget"); !reflect.DeepEqual(got, want) {
		t.Fatalf("LogTail = %v, want %v", got, want)
	}
	if msgs := collect(); len(msgs) != 0 {
		t.Fatalf("received %d snapshots while unsubscribed", len(msgs))
	}

	// Other managers have independent buffers.
	if got := relay.LogTail("Npm"); len(got) != 0 {
		t.Fatalf("unexpected log for Npm: %v", got)
	}
}

func TestRelayLogRingIsBounded(t *testing.T) {
	relay := NewProgressRelay()
	cb := relay.Callback("Winget")

	total := logRingSize + 50
	for i := 0; i < total; i++ {
		cb(fmt.Sprintf("line %d", i))
	}

	tail := relay.LogTail("Winget")
	if len(tail) != logRingSize {
		t.Fatalf("ring length = %d, want %d", len(tail), logRingSize)
	}
	if tail[0] != "line 50" {
		t.Fatalf("oldest retained line = %q, want line 50", tail[0])
	}
	if tail[len(tail)-1] != fmt.Sprintf("line %d", total-1) {
		t.Fatalf("newest retained line = %q", tail[len(tail)-1])
	}
}

func TestRelayCapsLogLineLength(t *testing.T) {
	relay := NewProgressRelay()
	cb := relay.Callback("Winget")

	cb(strings.Repeat("x", logLineCap*2))
	cb(strings.Repeat("y", logLineCap*2))

	tail := relay.LogTail("Winget")
	if len(tail) != 2 {
		t.Fatalf("got %d lines, want 2", len(tail))
	}
	for _, line := range tail {
		if len(line) != logLineCap {
			t.Fatalf("line length = %d, want %d", len(line), logLineCap)
		}
	}
}

func TestRelayEmitsSnapshotsWhenSubscribed(t *testing.T) {
	relay := NewProgressRelay()
	collect := collectLogMsgs(relay)
	relay.SubscribeLogs(true)

	cb := relay.Callback("Winget")
	cb("first")
	if msgs := collect(); len(msgs) != 1 {
		t.Fatalf("snapshots after first line = %d, want 1", len(msgs))
	}

	// Immediate follow-up is throttled but still retained.
	cb("second")
	if msgs := collect(); len(msgs) != 1 {
		t.Fatalf("throttle did not hold: %d snapshots", len(msgs))
	}
	if tail := relay.LogTail("Winget"); len(tail) != 2 {
		t.Fatalf("retained lines = %d, want 2", len(tail))
	}

	time.Sleep(logThrottle + 20*time.Millisecond)
	cb("third")

	msgs := collect()
	if len(msgs) != 2 {
		t.Fatalf("snapshots after throttle window = %d, want 2", len(msgs))
	}
	last := msgs[len(msgs)-1]
	if last.ManagerName != "Winget" {
		t.Fatalf("snapshot manager = %q", last.ManagerName)
	}
	if len(last.Lines) != 3 || last.Lines[2] != "third" {
		t.Fatalf("snapshot lines = %v", last.Lines)
	}

	// Unsubscribing stops snapshots.
	relay.SubscribeLogs(false)
	time.Sleep(logThrottle + 20*time.Millisecond)
	cb("fourth")
	if msgs := collect(); len(msgs) != 2 {
		t.Fatalf("snapshots after unsubscribe = %d, want 2", len(msgs))
	}
}

func TestRelayResetLog(t *testing.T) {
	relay := NewProgressRelay()
	cb := relay.Callback("Winget")
	cb("first")
	relay.ResetLog("Winget")
	if got := relay.LogTail("Winget"); len(got) != 0 {
		t.Fatalf("log not cleared: %v", got)
	}
}

// ---------------------------------------------------------------------------
// Model integration
// ---------------------------------------------------------------------------

func TestDetailsPaneToggleAndLogMessages(t *testing.T) {
	relay := NewProgressRelay()
	relay.Bind(func(tea.Msg) {})

	managers := []ecosystem.PackageManager{&fakeMgr{name: "Alpha"}}
	opts := testOptions(managers)
	opts.ProgressRelay = relay

	m := NewModel(context.Background(), opts)
	m = apply(t, m, scanDoneMsg{managers: managers})

	// Expand.
	m = apply(t, m, runeKey('d'))
	if !m.expanded {
		t.Fatal("pane did not expand on 'd'")
	}
	if m.logManager != "Alpha" {
		t.Fatalf("log manager = %q, want Alpha", m.logManager)
	}

	// Snapshots update the cached lines while expanded.
	m = apply(t, m, LogMsg{ManagerName: "Alpha", Lines: []string{"a", "b"}})
	if !reflect.DeepEqual(m.logLines, []string{"a", "b"}) {
		t.Fatalf("log lines = %v", m.logLines)
	}

	// Esc collapses without opening the cancel dialog.
	m = apply(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.expanded {
		t.Fatal("pane did not collapse on esc")
	}
	if m.state != stateUpdating {
		t.Fatalf("state = %v, want updating", m.state)
	}

	// A collapsed pane ignores snapshots.
	m = apply(t, m, LogMsg{ManagerName: "Alpha", Lines: []string{"c"}})
	if m.logLines != nil {
		t.Fatalf("collapsed pane accepted log lines: %v", m.logLines)
	}

	// Esc with the pane closed still opens the cancel dialog.
	m = apply(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.state != stateConfirmCancel {
		t.Fatalf("state = %v, want confirm-cancel", m.state)
	}
}

func TestDetailsPaneSeedsFromRelay(t *testing.T) {
	relay := NewProgressRelay()
	relay.Bind(func(tea.Msg) {})

	managers := []ecosystem.PackageManager{&fakeMgr{name: "Alpha"}}
	opts := testOptions(managers)
	opts.ProgressRelay = relay

	m := NewModel(context.Background(), opts)
	m = apply(t, m, scanDoneMsg{managers: managers})

	// Output arrives while the manager is running, before the pane is opened.
	cb := relay.Callback("Alpha")
	cb("already ran")

	m = apply(t, m, runeKey('d'))
	if !reflect.DeepEqual(m.logLines, []string{"already ran"}) {
		t.Fatalf("seeded log = %v, want the retained tail", m.logLines)
	}
}

func TestDetailsPaneShowsFailedManager(t *testing.T) {
	relay := NewProgressRelay()
	relay.Bind(func(tea.Msg) {})

	failing := &fakeMgr{name: "Alpha", fail: true}
	opts := testOptions([]ecosystem.PackageManager{failing, &fakeMgr{name: "Beta"}})
	opts.ProgressRelay = relay

	m := NewModel(context.Background(), opts)
	m = apply(t, m, scanDoneMsg{managers: opts.Managers})
	m = apply(t, m, runeKey('d'))
	m = apply(t, m, UpdateMsg{
		ManagerName: "Alpha",
		Result:      ecosystem.Result{Success: false, Message: "boom"},
	})

	if m.state != stateErrorPrompt {
		t.Fatalf("state = %v, want error prompt", m.state)
	}
	if got := m.activeManagerName(); got != "Alpha" {
		t.Fatalf("pane manager = %q, want the failed manager", got)
	}
}

func TestDetailsPaneRendering(t *testing.T) {
	relay := NewProgressRelay()
	relay.Bind(func(tea.Msg) {})

	managers := []ecosystem.PackageManager{&fakeMgr{name: "Alpha"}}
	opts := testOptions(managers)
	opts.ProgressRelay = relay

	m := NewModel(context.Background(), opts)
	m = apply(t, m, scanDoneMsg{managers: managers})
	m = apply(t, m, runeKey('d'))
	m = apply(t, m, LogMsg{ManagerName: "Alpha", Lines: []string{
		"Downloading alpha [45%]",
		"error: something failed badly",
		"installed alpha [1.0.0]",
	}})
	m.width, m.height = 100, 40

	out := ansi.Strip(m.render())
	if !strings.Contains(out, "Downloading alpha [45%]") {
		t.Fatalf("log pane missing captured output:\n%s", out)
	}
	if !strings.Contains(out, "error: something failed badly") {
		t.Fatalf("log pane missing error line:\n%s", out)
	}
	if strings.Contains(out, "waiting for output") {
		t.Fatalf("placeholder shown despite output:\n%s", out)
	}
}
