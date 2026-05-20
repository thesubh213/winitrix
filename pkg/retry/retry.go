package retry

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ComputeDelay returns a retry delay with optional exponential backoff and jitter.
func ComputeDelay(attempt int, base time.Duration, backoff bool, max time.Duration) time.Duration {
	if attempt <= 1 {
		return 0
	}

	delay := base
	if backoff {
		shift := attempt - 2
		if shift < 0 {
			shift = 0
		}
		delay = base * time.Duration(1<<shift)
	}

	if max > 0 && delay > max {
		delay = max
	}

	// Add up to 25% jitter to spread retries.
	if delay > 0 {
		jitter := time.Duration(rng.Int63n(int64(delay/4) + 1))
		delay += jitter
	}

	return delay
}
