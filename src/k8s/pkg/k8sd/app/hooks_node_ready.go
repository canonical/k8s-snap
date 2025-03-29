package app

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
)

// onNodeReady is called when the node is ready, right after the wait group is released.
// The node is ready if:
// - the microcluster database is accessible
// - the kubernetes endpoint is reachable.
// Note that this is not a microcluster hook, but a custom k8sd hook.
func (a *App) onNodeReady(ctx context.Context) error {
	log := log.FromContext(ctx).WithValues("hook", "onNodeReady")

	// Apply all custom CRDs on startup
	log.Info("Applying custom CRDs")
	if err := a.applyCustomCRDs(ctx); err != nil {
		return fmt.Errorf("failed to apply custom CRDs: %w", err)
	}

	return nil
}

func (a *App) applyCustomCRDs(ctx context.Context) error {
	log := log.FromContext(ctx).WithValues("startup", "applyCustomCRDs", "dir", a.snap.K8sCRDDir())

	isWorker, err := snaputil.IsWorker(a.snap)
	if err != nil {
		return fmt.Errorf("failed to check if node is a worker: %w", err)
	}
	if isWorker {
		log.V(1).Info("Skipping custom CRD application on worker node")
		return nil
	}

	k8sClient, err := a.snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := k8sClient.ApplyCRDs(ctx, a.snap.K8sCRDDir()); err != nil {
		return fmt.Errorf("failed to apply custom CRDs: %w", err)
	}

	return nil
}
