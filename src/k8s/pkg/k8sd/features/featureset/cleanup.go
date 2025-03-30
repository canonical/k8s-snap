package featureset

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
)

type cleanup struct {
	cleanupNetwork func(context.Context, snap.Snap) error
}

func (c *cleanup) CleanupNetwork(ctx context.Context, snap snap.Snap) error {
	return c.cleanupNetwork(ctx, snap)
}
