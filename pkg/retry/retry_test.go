package retry

import (
	"testing"
	"time"
)

func TestComputeDelayFirstAttemptIsZero(t *testing.T) {
	if d := ComputeDelay(1, time.Second, true, 30*time.Second); d != 0 {
		t.Fatalf("delay = %s, want 0", d)
	}
}

func TestComputeDelayJitterBounds(t *testing.T) {
	base := time.Second
	maxAllowed := base + base/4 + time.Nanosecond

	for i := 0; i < 100; i++ {
		d := ComputeDelay(2, base, false, 0)
		if d < base || d > maxAllowed {
			t.Fatalf("delay %s outside jitter bounds [%s, %s]", d, base, maxAllowed)
		}
	}
}

func TestComputeDelayClampedToMax(t *testing.T) {
	maxDelay := 10 * time.Second
	maxAllowed := maxDelay + maxDelay/4 + time.Nanosecond

	for i := 0; i < 100; i++ {
		d := ComputeDelay(6, time.Second, true, maxDelay)
		if d < maxDelay || d > maxAllowed {
			t.Fatalf("delay %s not clamped to max %s", d, maxDelay)
		}
	}
}

// TestComputeDelayNoOverflow guards against the exponential shift overflowing
// time.Duration into a nonsensical (possibly negative) value.
func TestComputeDelayNoOverflow(t *testing.T) {
	for _, attempt := range []int{10, 40, 80, 500, 100000} {
		d := ComputeDelay(attempt, time.Hour, true, 0)
		if d <= 0 {
			t.Fatalf("attempt %d produced non-positive delay %s", attempt, d)
		}
	}

	maxDelay := 30 * time.Minute
	maxAllowed := maxDelay + maxDelay/4 + time.Nanosecond
	for _, attempt := range []int{40, 500, 100000} {
		d := ComputeDelay(attempt, time.Second, true, maxDelay)
		if d < maxDelay || d > maxAllowed {
			t.Fatalf("attempt %d produced %s, want it clamped near %s", attempt, d, maxDelay)
		}
	}
}
