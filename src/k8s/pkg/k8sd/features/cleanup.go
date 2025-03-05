package features

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
)

type CleanupInterface interface {
	CleanupNetwork(context.Context, snap.Snap) error
}
