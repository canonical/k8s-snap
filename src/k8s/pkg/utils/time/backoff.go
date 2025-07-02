package time

import (
	"math/rand/v2"
	"time"
)

// ExponentialBackoff returns a duration that grows exponentially with attempt number,
// capped at maxDelay, and with ±50% jitter.
func ExponentialBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	exp := min(max(0, attempt), 32) // cap at 32 to avoid overflow
	// exponential: base * 2^exp
	d := min(baseDelay<<exp, maxDelay)
	// add jitter in range [d/2, 3d/2)
	jitter := time.Duration(rand.Int64N(int64(d))) - d/2
	return d + jitter
}
