package features

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
)

type StatusInterface interface {
	CheckDNS(context.Context, snap.Snap) (bool, error)
	CheckNetwork(context.Context, snap.Snap) (bool, error)
}

type statusChecks struct {
	checkDNS     func(context.Context, snap.Snap) (bool, error)
	checkNetwork func(context.Context, snap.Snap) (bool, error)
}

func (s *statusChecks) CheckDNS(ctx context.Context, snap snap.Snap) (bool, error) {
	return s.checkDNS(ctx, snap)
}

func (s *statusChecks) CheckNetwork(ctx context.Context, snap snap.Snap) (bool, error) {
	return s.checkNetwork(ctx, snap)
}
