package tui

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/thesubh213/winitrix/pkg/ecosystem"
	"github.com/thesubh213/winitrix/pkg/runner"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeMgr struct {
	name      string
	fail      bool
	delay     time.Duration
	progress  bool
	calls     int32
	cancelled *atomic.Bool
}

func (f *fakeMgr) Name() string { return f.name }

func (f *fakeMgr) Detect(ctx context.Context) bool { return true }

func (f *fakeMgr) UpdateAll(ctx context.Context) ecosystem.Result {
	return f.UpdateAllWithProgress(ctx, nil)
}

func (f *fakeMgr) UpdateAllWithProgress(ctx context.Context, cb runner.ProgressCallback) ecosystem.Result {
	atomic.AddInt32(&f.calls, 1)

	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			if f.cancelled != nil {
				f.cancelled.Store(true)
			}
			return ecosystem.Result{Success: false, Message: "cancelled by user", Error: ctx.Err()}
		}
	}

	if f.progress && cb != nil {
		cb("Downloading: alpha-package [1.0.0 42%]")
		cb("Installing: alpha-package [1.0.0 42%]")
	}

	if f.fail {
		return ecosystem.Result{Success: false, Message: "boom", Error: errors.New("exit status 1")}
	}
	return ecosystem.Result{Success: true, Message: "ok", Updated: 1}
}

func (f *fakeMgr) Clean(ctx context.Context) ecosystem.Result { return ecosystem.Result{Success: true} }
func (f *fakeMgr) Doctor(ctx context.Context) ecosystem.Result {
	return ecosystem.Result{Success: true}
}

type flakyMgr struct {
	fakeMgr
	attempts int32
}

func (f *flakyMgr) UpdateAllWithProgress(ctx context.Context, cb runner.ProgressCallback) ecosystem.Result {
	if atomic.AddInt32(&f.attempts, 1) == 1 {
		return ecosystem.Result{Success: false, Message: "transient failure", Error: errors.New("exit status 1")}
	}
	return ecosystem.Result{Success: true, Message: "ok on retry", Updated: 1}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func apply(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, _ := m.Update(msg)
	model, ok := next.(Model)
	if !ok {
		t.Fatalf("Update returned %T, want Model", next)
	}
	return model
}

func runeKey(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func ctrlC() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
}

func testOptions(managers []ecosystem.PackageManager) Options {
	return Options{
		Managers:       managers,
		TimeoutMinutes: 1,
		Animations:     false,
		AltScreen:      false,
	}
}

func manyManagers(n int) []ecosystem.PackageManager {
	managers := make([]ecosystem.PackageManager, 0, n)
	for i := 0; i < n; i++ {
		managers = append(managers, &fakeMgr{name: "Manager" + strings.Repeat("X", i%5) + string(rune('A'+i%26))})
	}
	return managers
}

// ---------------------------------------------------------------------------
// Render invariants
// ---------------------------------------------------------------------------

// TestRenderNeverExceedsWidth is the core layout guard: for every state and a
// wide range of terminal sizes, no rendered line may exceed the screen width.
func TestRenderNeverExceedsWidth(t *testing.T) {
	managers := manyManagers(25)

	buildStates := func() map[string]Model {
		states := map[string]Model{}

		// Scanning state.
		scanning := NewModel(context.Background(), testOptions(managers))
		states["scanning"] = apply(t, scanning, tea.WindowSizeMsg{Width: 100, Height: 40})

		// Selection state.
		selectOpts := testOptions(managers)
		selectOpts.SelectManagers = true
		sel := NewModel(context.Background(), selectOpts)
		sel = apply(t, sel, scanDoneMsg{managers: managers})
		states["selection"] = sel

		// Updating state with one finished manager, one active with progress.
		updating := NewModel(context.Background(), testOptions(managers))
		updating = apply(t, updating, scanDoneMsg{managers: managers})
		updating = apply(t, updating, UpdateMsg{
			ManagerName: managers[0].Name(),
			Result:      ecosystem.Result{Success: true, Message: "ok", Updated: 3},
			Duration:    2 * time.Second,
		})
		updating = apply(t, updating, ProgressMsg{
			ManagerName: managers[1].Name(),
			Status:      "Installing",
			PackageName: "very-long-package-name-with-suffix",
			Percentage:  42,
		})
		states["updating"] = updating

		// Error prompt.
		errModel := NewModel(context.Background(), testOptions(managers))
		errModel = apply(t, errModel, scanDoneMsg{managers: managers})
		errModel = apply(t, errModel, UpdateMsg{
			ManagerName: managers[0].Name(),
			Result: ecosystem.Result{
				Success: false,
				Message: "Failed to update: something went wrong with a long explanatory message",
				Error:   errors.New("exit status 1: detailed underlying error text"),
			},
			Duration: time.Second,
		})
		states["error"] = errModel

		// Expanded details pane, with and without captured output.
		expanded := updating
		expanded = apply(t, expanded, runeKey('d'))
		expanded = apply(t, expanded, LogMsg{
			ManagerName: managers[1].Name(),
			Lines: []string{
				"Downloading alpha [45%]",
				"error: something failed with a very long explanatory message that must be truncated to the pane width",
				strings.Repeat("wide ", 80),
			},
		})
		states["updating-expanded"] = expanded

		errExpanded := errModel
		errExpanded = apply(t, errExpanded, runeKey('d'))
		errExpanded = apply(t, errExpanded, LogMsg{ManagerName: managers[0].Name(), Lines: []string{"line one", "line two"}})
		states["error-expanded"] = errExpanded

		// Cancel confirmation.
		cancel := NewModel(context.Background(), testOptions(managers))
		cancel = apply(t, cancel, scanDoneMsg{managers: managers})
		cancel = apply(t, cancel, ctrlC())
		states["cancel"] = cancel

		// Done with mixed results (nonstop so the failure does not pause).
		doneManagers := manyManagers(2)
		doneOpts := testOptions(doneManagers)
		doneOpts.NonStop = true
		done := NewModel(context.Background(), doneOpts)
		done = apply(t, done, scanDoneMsg{managers: doneManagers})
		done = apply(t, done, UpdateMsg{ManagerName: doneManagers[0].Name(), Result: ecosystem.Result{Success: true, Updated: 2}, Duration: time.Second})
		done = apply(t, done, UpdateMsg{ManagerName: doneManagers[1].Name(), Result: ecosystem.Result{Success: false, Message: "nope"}, Duration: time.Second})
		states["done"] = done

		// No managers detected.
		empty := NewModel(context.Background(), testOptions(nil))
		empty = apply(t, empty, scanDoneMsg{managers: nil})
		states["empty"] = empty

		return states
	}

	for name, m := range buildStates() {
		for _, width := range []int{20, 30, 40, 61, 80, 100, 120, 200} {
			for _, height := range []int{8, 12, 24, 40, 60} {
				sized := m
				sized.width = width
				sized.height = height

				out := sized.render()
				for _, line := range strings.Split(out, "\n") {
					if w := ansi.StringWidth(line); w > width {
						t.Fatalf("state %s at %dx%d: line width %d exceeds width: %q",
							name, width, height, w, ansi.Strip(line))
					}
				}
			}
		}
	}
}

func TestModelResultsBoundsGuard(t *testing.T) {
	managers := []ecosystem.PackageManager{&fakeMgr{name: "Alpha"}}
	m := NewModel(context.Background(), testOptions(managers))
	m = apply(t, m, scanDoneMsg{managers: managers})

	// A stray progress message for an unknown manager must not corrupt state.
	m = apply(t, m, ProgressMsg{ManagerName: "NotRunning", Status: "x"})
	if m.currentManagerIdx != 0 {
		t.Fatalf("currentManagerIdx = %d, want 0", m.currentManagerIdx)
	}

	// Completion for a manager that never started must not panic the view.
	m = apply(t, m, UpdateMsg{ManagerName: "Alpha", Result: ecosystem.Result{Success: true}, Duration: time.Second})
	if len(m.results) != 1 {
		t.Fatalf("results = %d, want 1", len(m.results))
	}
	_ = m.render()
}

// ---------------------------------------------------------------------------
// teatest flows
// ---------------------------------------------------------------------------

func TestModelHappyPath(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Alpha", progress: true, delay: 20 * time.Millisecond},
		&fakeMgr{name: "Beta"},
	}
	m := NewModel(context.Background(), testOptions(managers))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final, ok := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if !ok {
		t.Fatal("final model has unexpected type")
	}
	results := final.GetResults()
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	for _, res := range results {
		if !res.Result.Success {
			t.Fatalf("manager %s failed unexpectedly: %s", res.ManagerName, res.Result.Message)
		}
	}
}

// streamingFake emits output with a gap so frames actually get flushed
// between messages (Bubble Tea coalesces frames that arrive together).
type streamingFake struct {
	name       string
	firstLine  string
	secondLine string
	hold       time.Duration
}

func (f *streamingFake) Name() string { return f.name }

func (f *streamingFake) Detect(ctx context.Context) bool { return true }

func (f *streamingFake) UpdateAll(ctx context.Context) ecosystem.Result {
	return f.UpdateAllWithProgress(ctx, nil)
}

func (f *streamingFake) UpdateAllWithProgress(ctx context.Context, cb runner.ProgressCallback) ecosystem.Result {
	if cb != nil {
		cb(f.firstLine)
	}
	select {
	case <-time.After(f.hold):
	case <-ctx.Done():
		return ecosystem.Result{Success: false, Message: "cancelled", Error: ctx.Err()}
	}
	if cb != nil {
		cb(f.secondLine)
	}
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return ecosystem.Result{Success: false, Message: "cancelled", Error: ctx.Err()}
	}
	return ecosystem.Result{Success: true, Message: "ok", Updated: 1}
}

func (f *streamingFake) Clean(ctx context.Context) ecosystem.Result {
	return ecosystem.Result{Success: true}
}
func (f *streamingFake) Doctor(ctx context.Context) ecosystem.Result {
	return ecosystem.Result{Success: true}
}

func TestModelDetailsPaneFlow(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&streamingFake{
			name:       "Alpha",
			firstLine:  "Downloading: alpha-package [1.0.0 20%]",
			secondLine: "Installing: alpha-package [1.0.0 90%]",
			hold:       400 * time.Millisecond,
		},
		&fakeMgr{name: "Beta"},
	}
	relay := NewProgressRelay()
	opts := testOptions(managers)
	opts.ProgressRelay = relay

	m := NewModel(context.Background(), opts)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))
	relay.Bind(tm.Send)

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Alpha"))
	}, teatest.WithDuration(15*time.Second))

	// Expand the pane: the retained first line appears immediately.
	tm.Send(runeKey('d'))
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Downloading: alpha-package"))
	}, teatest.WithDuration(15*time.Second))

	// Later output streams in live while the pane is open.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Installing: alpha-package"))
	}, teatest.WithDuration(15*time.Second))

	// Collapsing must not interrupt the run.
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEscape})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if got := len(final.GetResults()); got != 2 {
		t.Fatalf("results = %d, want 2", got)
	}
}

func TestModelErrorPromptSkip(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Broken", fail: true},
		&fakeMgr{name: "Fine"},
	}
	m := NewModel(context.Background(), testOptions(managers))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("retry"))
	}, teatest.WithDuration(15*time.Second))

	// Skip the failed manager.
	tm.Send(runeKey('s'))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	results := final.GetResults()
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Result.Success {
		t.Fatal("the failed manager must be recorded as a failure")
	}
}

func TestModelErrorPromptRetry(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&flakyMgr{fakeMgr: fakeMgr{name: "Flaky"}},
		&fakeMgr{name: "Fine"},
	}
	m := NewModel(context.Background(), testOptions(managers))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("retry"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(runeKey('r'))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	results := final.GetResults()
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2 (retry replaces the failure)", len(results))
	}
	if !results[0].Result.Success {
		t.Fatalf("retried manager should succeed, got %q", results[0].Result.Message)
	}
}

func TestModelCancelDeclined(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Slow", delay: 250 * time.Millisecond},
		&fakeMgr{name: "Fine"},
	}
	m := NewModel(context.Background(), testOptions(managers))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Slow"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(ctrlC())
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Cancel package updates"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(runeKey('n'))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if got := len(final.GetResults()); got != 2 {
		t.Fatalf("results = %d, want both managers to finish", got)
	}
}

func TestModelCancelConfirmed(t *testing.T) {
	var ctxCancelled atomic.Bool
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Slow", delay: 2 * time.Second, cancelled: &ctxCancelled},
		&fakeMgr{name: "Fine"},
	}
	m := NewModel(context.Background(), testOptions(managers))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Slow"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(ctrlC())
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Cancel package updates"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(runeKey('y'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(10*time.Second))

	// The running manager must observe the cancellation (in production this is
	// what terminates the package-manager process tree).
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && !ctxCancelled.Load() {
		time.Sleep(20 * time.Millisecond)
	}
	if !ctxCancelled.Load() {
		t.Fatal("the running manager's context was not cancelled")
	}

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	results := final.GetResults()
	if len(results) == 0 {
		t.Fatal("cancelled run must still report the interrupted manager")
	}
	last := results[len(results)-1]
	if last.Result.Success || !strings.Contains(last.Result.Message, "cancelled") {
		t.Fatalf("last result = %+v, want a cancellation marker", last.Result)
	}
}

func TestModelResizeDuringRun(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Alpha", delay: 150 * time.Millisecond, progress: true},
		&fakeMgr{name: "Beta"},
	}
	m := NewModel(context.Background(), testOptions(managers))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Alpha"))
	}, teatest.WithDuration(15*time.Second))

	// Shrink and grow mid-run; the model must keep rendering.
	tm.Send(tea.WindowSizeMsg{Width: 40, Height: 12})
	tm.Send(tea.WindowSizeMsg{Width: 200, Height: 60})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if got := len(final.GetResults()); got != 2 {
		t.Fatalf("results = %d, want 2", got)
	}
}

func TestViewDeclaresAltScreenAndProgress(t *testing.T) {
	managers := []ecosystem.PackageManager{&fakeMgr{name: "Alpha"}}
	opts := testOptions(managers)
	opts.AltScreen = true

	m := NewModel(context.Background(), opts)
	v := m.View()
	if !v.AltScreen {
		t.Fatal("View must request the alternate screen when configured")
	}
	if v.WindowTitle != "winitrix" {
		t.Fatalf("window title = %q", v.WindowTitle)
	}

	m = apply(t, m, scanDoneMsg{managers: managers})
	v = m.View()
	if v.ProgressBar == nil {
		t.Fatal("updating view should publish a native terminal progress bar")
	}
}

func TestModelAnimatedFlow(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Alpha", progress: true, delay: 30 * time.Millisecond},
		&fakeMgr{name: "Beta"},
	}
	opts := testOptions(managers)
	opts.Animations = true

	m := NewModel(context.Background(), opts)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if got := len(final.GetResults()); got != 2 {
		t.Fatalf("results = %d, want 2", got)
	}
}

func TestModelSelectionFlow(t *testing.T) {
	managers := []ecosystem.PackageManager{
		&fakeMgr{name: "Alpha"},
		&fakeMgr{name: "Beta"},
	}
	opts := testOptions(managers)
	opts.SelectManagers = true

	m := NewModel(context.Background(), opts)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Select ecosystems"))
	}, teatest.WithDuration(15*time.Second))

	// Deselect all, then select only the second manager, then start.
	tm.Send(runeKey('n'))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyDown})
	tm.Send(runeKey(' '))
	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Done"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	results := final.GetResults()
	if len(results) != 1 || results[0].ManagerName != "Beta" {
		t.Fatalf("results = %+v, want only Beta", results)
	}
}

func TestModelQuitFromSelection(t *testing.T) {
	managers := []ecosystem.PackageManager{&fakeMgr{name: "Alpha"}}
	opts := testOptions(managers)
	opts.SelectManagers = true

	m := NewModel(context.Background(), opts)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 40))

	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Select ecosystems"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(runeKey('q'))
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("no updates were run"))
	}, teatest.WithDuration(15*time.Second))

	tm.Send(tea.KeyPressMsg{Code: tea.KeyEnter})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))

	final := tm.FinalModel(t, teatest.WithFinalTimeout(5*time.Second)).(Model)
	if got := len(final.GetResults()); got != 0 {
		t.Fatalf("results = %d, want 0", got)
	}
}
