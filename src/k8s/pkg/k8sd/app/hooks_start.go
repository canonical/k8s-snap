package app

import (
	"context"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/state"
)

func (a *App) onStart(s *state.State) error {
	// start a goroutine to mark the node as running
	go a.markNodeReady(s.Context, s)

	// start node config controller
	if a.nodeConfigController != nil {
		go a.nodeConfigController.Run(s.Context)
	}

	// start control plane config controller
	if a.controlPlaneConfigController != nil {
		go a.controlPlaneConfigController.Run(s.Context, func(ctx context.Context) (types.ClusterConfig, error) {
			return utils.GetClusterConfig(ctx, s)
		})
	}

	return nil
}
