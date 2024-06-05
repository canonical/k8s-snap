package fake

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyNetwork is a dummy implementation of the ApplyNetwork function. It does nothing and returns nil.
func ApplyNetwork(ctx context.Context, snap snap.Snap, cfg types.Network, _ types.Annotations) error {
	return nil
}

// CheckNetwork is a dummy implementation of the CheckNetwork function. It does nothing and returns true.
func CheckNetwork(ctx context.Context, snap snap.Snap) (bool, error) {
	return true, nil
}
