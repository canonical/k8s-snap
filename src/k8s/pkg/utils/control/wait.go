package control

import (
	"context"
	"fmt"
	"time"
)

// WaitUntilReady waits until the specified condition becomes true.
// checkFunc can return an error to return early.
func WaitUntilReady(ctx context.Context, checkFunc func() (bool, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			if ok, err := checkFunc(); err != nil {
				return fmt.Errorf("wait check failed: %w", err)
			} else if ok {
				return nil
			}
		}
	}
}
