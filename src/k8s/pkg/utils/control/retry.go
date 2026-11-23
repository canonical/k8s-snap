package control

import (
	"context"
	"time"
)

// RetryFor will retry a given function for the given amount of times.
// RetryFor will wait for backoff between retries.
func RetryFor(ctx context.Context, retryCount int, delayBetweenRetry time.Duration, retryFunc func() error) error {
	return RetryForIf(ctx, retryCount, delayBetweenRetry, func(error) bool { return true }, retryFunc)
}

// RetryForIf behaves like RetryFor, but only retries when isRetryable(err)
// returns true for the error retryFunc just returned. Any other error is
// returned immediately, without waiting out the remaining retry budget.
func RetryForIf(ctx context.Context, retryCount int, delayBetweenRetry time.Duration, isRetryable func(error) bool, retryFunc func() error) error {
	var err error = nil
	for range retryCount {
		if err = retryFunc(); err != nil {
			if !isRetryable(err) {
				return err
			}
			select {
			case <-ctx.Done():
				return context.Canceled
			case <-time.After(delayBetweenRetry):
				continue
			}
		}
		break
	}
	return err
}
