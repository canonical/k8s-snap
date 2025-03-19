package app

import (
	"context"
	"fmt"

	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils/experimental/snapdconfig"
	"github.com/canonical/microcluster/v2/state"
)

func (a *App) postRefreshHook(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("hook", "post-refresh")
	log.Info("Running post-refresh hook")

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
