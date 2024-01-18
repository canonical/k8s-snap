package utils

import (
	"context"
	"fmt"
	"time"
)

// WaitUntilReady waits until the specified condition becomes true.
func WaitUntilReady(ctx context.Context, checkFunc func() bool, timeout time.Duration, errorMessage string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout):
			return fmt.Errorf("%s: %w", errorMessage, context.DeadlineExceeded)
		default:
			ready := checkFunc()
			if ready {
				return nil
			}
			<-time.After(time.Second)
		}
	}
}
