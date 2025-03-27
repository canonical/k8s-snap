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
