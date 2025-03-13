package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/features/manifests"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

// FeatureManifestController is a controller that syncs feature manifests available in the snap with the microcluster database.
type FeatureManifestController struct {
	snap      snap.Snap
	waitReady func()
}

// NewFeatureManifestController creates a new feature manifest controller.
func NewFeatureManifestController(snap snap.Snap, waitReady func()) *FeatureManifestController {
	return &FeatureManifestController{
		snap:      snap,
		waitReady: waitReady,
	}
}

// Run runs the feature manifest controller.
// The controller reconciles feature manifests available in the snap with the microcluster database until all manifests are inserted.
func (c *FeatureManifestController) Run(ctx context.Context, s state.State) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "feature-manifest"))
	log := log.FromContext(ctx)

	log.Info("Waiting for node to be ready")
	// wait for microcluster node to be ready
	c.waitReady()

	log.Info("Starting feature manifest controller")

	for {
		log.Info("Reconciling feature manifests")

		var retryRequired bool

		for _, featureManifest := range manifests.Manifests() {
			log := log.WithValues("feature", featureManifest.GetName(), "version", featureManifest.GetVersion())

			if err := c.reconcile(ctx, s, featureManifest); err != nil {
				log.Error(err, "failed to reconcile feature manifest")
				retryRequired = true
			}
		}

		if !retryRequired {
			log.Info("Reconcilation of feature manifests complete")
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *FeatureManifestController) reconcile(ctx context.Context, s state.State, manifest *types.FeatureManifest) error {
	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return database.InsertFeatureManifest(ctx, tx, manifest)
	}); err != nil {
		return fmt.Errorf("failed to insert feature manifest: %w", err)
	}

	return nil
}
