package internal

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/microcluster/v2/state"
)

type updateClusterDNSFunc func(ctx context.Context, s state.State, dnsIP string) error

var UpdateClusterDNS updateClusterDNSFunc = features.UpdateClusterDNS

func MockUpdateClusterDNS(ctx context.Context, s state.State, dnsIP string) error {
	return nil
}
