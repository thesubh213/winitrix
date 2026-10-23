package retry

import (
	"math"
	"math/rand"
	"time"
)

// maxBackoffShift caps the exponential backoff shift so that the delay
// computation cannot overflow a time.Duration.
const maxBackoffShift = 30

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
		if shift > maxBackoffShift {
			shift = maxBackoffShift
		}

		// Guard the shift against overflow instead of trusting the result.
		if shift > 0 && base > time.Duration(math.MaxInt64>>uint(shift)) {
			delay = max
			if delay <= 0 {
				delay = base
			}
		} else {
			delay = base << uint(shift)
		}
	}

	if max > 0 && delay > max {
		delay = max
	}

	// Add up to 25% jitter to spread retries. The top-level math/rand
	// functions are safe for concurrent use.
	if delay > 0 {
		jitter := time.Duration(rand.Int63n(int64(delay/4) + 1))
		delay += jitter
	}

	return delay
}
