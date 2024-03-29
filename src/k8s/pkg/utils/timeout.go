package utils

import (
	"context"
	"time"
)

func TimeoutFromCtx(ctx context.Context, defaultTimeout time.Duration) time.Duration {
	timeout := defaultTimeout

	if deadline, set := ctx.Deadline(); set {
		timeout = time.Until(deadline)
	}
	return timeout
}
