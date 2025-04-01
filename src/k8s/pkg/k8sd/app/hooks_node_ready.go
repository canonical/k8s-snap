package app

import (
	"context"
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
)

// onNodeReady is called when the node is ready, right after the wait group is released.
// The node is ready if:
// - the microcluster database is accessible
// - the kubernetes endpoint is reachable.
// Note that this is not a microcluster hook, but a custom k8sd hook.
func (a *App) onNodeReady(ctx context.Context, s state.State) error {
	log := log.FromContext(ctx).WithValues("hook", "onNodeReady")

	// Check if a refresh was performed and if so, run the custom post-refresh hook
	log.Info("Checking if snap is post-refresh")
	isPostRefresh, err := utils.FileExists(a.snap.PostRefreshLockPath())
	if err != nil {
		return fmt.Errorf("failed to check if snap is post-refresh: %w", err)
	}
	if isPostRefresh {
		log.Info("Snap is post-refresh - running post-refresh hook")
		if err := a.postRefreshHook(ctx, s); err != nil {
			return fmt.Errorf("failed to run post-refresh hook: %w", err)
		}

		log.Info("Post-refresh hook completed successfully - removing lock file.")
		if err := os.Remove(a.snap.PostRefreshLockPath()); err != nil {
			return fmt.Errorf("failed to remove post-refresh lock file: %w", err)
		}
	} else {
		log.Info("Snap is not post-refresh")
	}

	return nil
}
