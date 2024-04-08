package control

import "context"

// RetryFor will retry a given function for the given amount of times.
// RetryFor will not wait between retries. This is up to the retryFunc to handle.
func RetryFor(ctx context.Context, retryCount int, retryFunc func() error) error {
	var err error = nil
	for i := 0; i < retryCount; i++ {
		if err = retryFunc(); err != nil {
			select {
			case <-ctx.Done():
				return context.Canceled
			default:
				continue
			}
		}
		break
	}
	return err
}
