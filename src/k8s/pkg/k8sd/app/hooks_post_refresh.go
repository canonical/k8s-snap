package app

import (
	"context"
	"fmt"

	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/log"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/v2/state"
)

// postRefreshHook is executed after the node is ready after a `snap refresh` operation
// See nodeReadyHook for details on when a node is considered ready.
// Note that the postRefreshHook is NOT executed after a `snap install` operation which is
// different to the underlying snap hook.
func (a *App) postRefreshHook(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("hook", "post-refresh")
	log.Info("Running post-refresh hook")

	isWorker, err := snaputil.IsWorker(a.snap)
	if err != nil {
		return fmt.Errorf("failed to check if node is a worker: %w", err)
	}

	if isWorker {
		log.Info("Node is a worker, skipping post-refresh hook")
		return nil
	}

	config, err := databaseutil.GetClusterConfig(ctx, s)
	if err != nil {
		return fmt.Errorf("failed to get cluster config: %w", err)
	}

	log.Info("Re-enable snapd/k8sd config sync and reconcile.")
	if err := snapdconfig.SetSnapdFromK8sd(ctx, config.ToUserFacing(), a.snap); err != nil {
		return fmt.Errorf("failed to set snapd configuration from k8sd: %w", err)
	}

	return nil
}
