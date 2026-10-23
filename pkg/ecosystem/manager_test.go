package ecosystem

import (
	"context"
	"testing"
	"time"

	"github.com/thesubh213/winitrix/pkg/runner"
)

type fakeProbe struct {
	name      string
	installed bool
	delay     time.Duration
	hang      bool
	probed    int
}

func (f *fakeProbe) Name() string { return f.name }

func (f *fakeProbe) Detect(ctx context.Context) bool {
	f.probed++
	if f.hang {
		<-ctx.Done()
		return false
	}
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return false
		}
	}
	return f.installed
}

func (f *fakeProbe) UpdateAll(ctx context.Context) Result {
	return Result{Success: true}
}

func (f *fakeProbe) UpdateAllWithProgress(ctx context.Context, cb runner.ProgressCallback) Result {
	return Result{Success: true}
}

func (f *fakeProbe) Clean(ctx context.Context) Result  { return Result{Success: true} }
func (f *fakeProbe) Doctor(ctx context.Context) Result { return Result{Success: true} }

func TestDetectManagersPreservesOrderAndFilters(t *testing.T) {
	slow := &fakeProbe{name: "SlowProbe", installed: true, delay: 60 * time.Millisecond}
	fast := &fakeProbe{name: "FastProbe", installed: true}
	missing := &fakeProbe{name: "MissingProbe", installed: false}

	got := DetectManagers(context.Background(), []PackageManager{slow, fast, missing}, nil)
	if len(got) != 2 {
		t.Fatalf("detected %d managers, want 2", len(got))
	}
	if got[0].Name() != "SlowProbe" || got[1].Name() != "FastProbe" {
		t.Fatalf("order not preserved: %s, %s", got[0].Name(), got[1].Name())
	}
}

func TestDetectManagersReportsProgress(t *testing.T) {
	managers := []PackageManager{
		&fakeProbe{name: "A", installed: true},
		&fakeProbe{name: "B", installed: false},
		&fakeProbe{name: "C", installed: true},
	}

	var lastProbed, lastTotal int
	found := 0
	DetectManagers(context.Background(), managers, func(name string, probed, total int) {
		if name != "" {
			found++
		}
		lastProbed = probed
		lastTotal = total
	})

	if lastTotal != 3 || lastProbed != 3 {
		t.Fatalf("final progress = %d/%d, want 3/3", lastProbed, lastTotal)
	}
	if found != 2 {
		t.Fatalf("found events = %d, want 2", found)
	}
}

func TestDetectManagersTimesOutHungProbe(t *testing.T) {
	oldTimeout := detectTimeout
	detectTimeout = 100 * time.Millisecond
	defer func() { detectTimeout = oldTimeout }()

	hung := &fakeProbe{name: "HungProbe", hang: true}
	ok := &fakeProbe{name: "OkProbe", installed: true}

	start := time.Now()
	got := DetectManagers(context.Background(), []PackageManager{hung, ok}, nil)
	elapsed := time.Since(start)

	if elapsed > 3*time.Second {
		t.Fatalf("hung probe was not bounded: %s", elapsed)
	}
	if len(got) != 1 || got[0].Name() != "OkProbe" {
		t.Fatalf("got %v, want just OkProbe", got)
	}
}

func TestBinaryAvailableForUnknownManagerProbes(t *testing.T) {
	// Managers without candidate binaries are always probed.
	if !binaryAvailable("Not A Real Manager") {
		t.Fatal("unknown managers must not be filtered out")
	}
}

func TestManagerBinariesCoverAllManagers(t *testing.T) {
	for _, m := range GetAllManagers() {
		if len(managerBinaries[m.Name()]) == 0 {
			t.Errorf("manager %q has no binary prefilter entry", m.Name())
		}
	}
}
