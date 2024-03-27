package utils

import (
	"context"
	"time"
)

func TimeoutFromCtx(ctx context.Context) time.Duration {
	timeout := 30 * time.Second
	if deadline, set := ctx.Deadline(); set {
		timeout = time.Until(deadline)
	}
	return timeout
}
