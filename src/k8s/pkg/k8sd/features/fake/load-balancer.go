package fake

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

// ApplyLoadBalancer is a dummy implementation of the ApplyLoadBalancer function. It does nothing and returns nil.
func ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network, _ types.Annotations) error {
	return nil
}
