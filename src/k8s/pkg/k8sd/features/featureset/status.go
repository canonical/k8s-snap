package featureset

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
)

// statusChecks implements the StatusInterface.
type statusChecks struct {
	checkDNS     func(context.Context, snap.Snap) error
	checkNetwork func(context.Context, snap.Snap) error
}

func (s *statusChecks) CheckDNS(ctx context.Context, snap snap.Snap) error {
	return s.checkDNS(ctx, snap)
}

func (s *statusChecks) CheckNetwork(ctx context.Context, snap snap.Snap) error {
	return s.checkNetwork(ctx, snap)
}
