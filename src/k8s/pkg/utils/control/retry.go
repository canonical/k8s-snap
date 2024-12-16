package control

import (
	"context"
	"time"
)

// RetryFor will retry a given function for the given amount of times.
// RetryFor will wait for backoff between retries.
func RetryFor(ctx context.Context, retryCount int, delayBetweenRetry time.Duration, retryFunc func() error) error {
	var err error = nil
	for i := 0; i < retryCount; i++ {
		if err = retryFunc(); err != nil {
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

// Consistently will retry a given function for the given amount of times, expecting all calls to succeed without
// errors. It will wait for backoff between retries.
func Consistently(ctx context.Context, retryCount int, delayBetweenRetry time.Duration, retryFunc func() error) error {
	for i := 0; i < retryCount; i++ {
		if err := retryFunc(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-time.After(delayBetweenRetry):
			continue
		}
	}
	return nil
}
