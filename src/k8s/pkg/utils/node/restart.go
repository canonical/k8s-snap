package node

import (
	"context"
	"time"

	"github.com/canonical/k8s/pkg/log"
)

// StartAsyncRestart initiates an asynchronous service restart process (defined
// as restartFn) after receiving a ready signal. It returns a channel that
// should be signaled when the operation is complete and ready for restart.
// The caller must close the channel after sending the signal.
func StartAsyncRestart(logger log.Logger, restartFn func(context.Context) error) chan error {
	readyCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		select {
		case err := <-readyCh:
			if err != nil {
				logger.Error(err, "Operation failed before restart")
				return
			}
		case <-ctx.Done():
			logger.Error(ctx.Err(), "Timeout waiting for operation")
			return
		}

		if err := restartFn(ctx); err != nil {
			logger.Error(err, "Failed to restart services")
		}
	}()
	return readyCh
}
