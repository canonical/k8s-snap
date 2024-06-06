package features

import (
	"context"

	"github.com/canonical/k8s/pkg/snap"
)

// StatusInterface defines the interface for checking the status of the built-in features.
type StatusInterface interface {
	// CheckDNS checks the status of the DNS feature.
	CheckDNS(context.Context, snap.Snap) error
	// CheckNetwork checks the status of the Network feature.
	CheckNetwork(context.Context, snap.Snap) error
}

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
