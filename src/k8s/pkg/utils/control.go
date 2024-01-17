package utils

import (
	"context"
	"fmt"
	"time"
)

// WaitUntilReady waits until the specified condition becomes true.
func WaitUntilReady(ctx context.Context, checkFunc func() bool, timeout time.Duration, errorMessage string) error {
	ch := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ready := checkFunc()
				if ready {
					ch <- struct{}{}
					return
				}
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(timeout):
		return fmt.Errorf("%s: %w", errorMessage, context.DeadlineExceeded)
	case <-ch:
		return nil
	}
}
