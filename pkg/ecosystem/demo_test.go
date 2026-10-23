package ecosystem

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDemoManagersShape(t *testing.T) {
	managers := NewDemoManagers()
	if len(managers) < 4 {
		t.Fatalf("expected several demo managers, got %d", len(managers))
	}

	sawFailure := false
	for _, m := range managers {
		if m.Name() == "" {
			t.Fatal("demo manager with empty name")
		}
		if !m.Detect(context.Background()) {
			t.Fatalf("demo manager %q must always detect", m.Name())
		}
		if res := m.Doctor(context.Background()); !res.Success {
			t.Fatalf("demo doctor for %q failed", m.Name())
		}
		if res := m.Clean(context.Background()); !res.Success {
			t.Fatalf("demo clean for %q failed", m.Name())
		}
		if m.(*demoManager).failAfter > 0 {
			sawFailure = true
		}
	}
	if !sawFailure {
		t.Fatal("demo set should include a failing manager so the error UI is reachable")
	}
}

func TestDemoManagerProgressAndFailure(t *testing.T) {
	var lines int
	failing := &demoManager{
		name:        "Failing",
		packages:    2,
		stepDelay:   40 * time.Millisecond,
		failAfter:   1,
		failMessage: "boom",
		failErr:     errors.New("exit status 1"),
	}

	res := failing.UpdateAllWithProgress(context.Background(), func(string) { lines++ })
	if res.Success {
		t.Fatal("expected failure result")
	}
	if res.Message != "boom" {
		t.Fatalf("message = %q, want boom", res.Message)
	}
	if lines == 0 {
		t.Fatal("expected progress callbacks before the failure")
	}
}

func TestDemoManagerSuccessAndCancel(t *testing.T) {
	ok := &demoManager{name: "Ok", packages: 1, stepDelay: 20 * time.Millisecond, updated: 1}
	res := ok.UpdateAllWithProgress(context.Background(), nil)
	if !res.Success || res.Updated != 1 {
		t.Fatalf("got %+v, want success with 1 update", res)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res = ok.UpdateAllWithProgress(ctx, nil)
	if res.Success {
		t.Fatal("cancelled context must not report success")
	}
	if !errors.Is(res.Error, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", res.Error)
	}
}
