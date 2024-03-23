package client

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/client"
	"github.com/canonical/microcluster/microcluster"
)

// k8sdClient interacts with the k8sd REST-API via unix-socket.
type k8sdClient struct {
	m    *microcluster.MicroCluster
	mc   *client.Client
	snap snap.Snap
}

// New returns a client to interact with the k8sd REST-API.
func New(ctx context.Context, snap snap.Snap) (*k8sdClient, error) {
	if snap == nil {
		panic("snap must not be nil")
	}
	m, err := microcluster.App(ctx, microcluster.Args{
		StateDir: snap.K8sdStateDir(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize microcluster app: %w", err)
	}
	mc, err := m.LocalClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create local microcluster client: %w", err)
	}

	return &k8sdClient{
		snap: snap,
		m:    m,
		mc:   mc,
	}, nil
}
