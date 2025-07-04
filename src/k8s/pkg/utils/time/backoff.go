package time

import (
	"math/rand/v2"
	"time"
)

var randInt64N = rand.Int64N

// LinearBackoff returns a duration that grows linearly with attempt number,
// capped at maxDelay, and with ±50% jitter.
func LinearBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	d := min(baseDelay, maxDelay)
	if attempt > 0 {
		d = min(baseDelay*time.Duration(attempt), maxDelay)
	}
	// add jitter in range [d/2, 3d/2)
	jitter := time.Duration(randInt64N(int64(d))) - d/2
	return d + jitter
}
