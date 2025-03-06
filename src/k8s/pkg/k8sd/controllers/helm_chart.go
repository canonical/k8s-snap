package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/charts"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// HelmChartController is a controller that syncs helm charts available in the snap with the microcluster database.
type HelmChartController struct {
	snap      snap.Snap
	waitReady func()
}

// NewHelmChartController creates a new helm chart controller.
func NewHelmChartController(snap snap.Snap, waitReady func()) *HelmChartController {
	return &HelmChartController{
		snap:      snap,
		waitReady: waitReady,
	}
}

// Run runs the helm chart controller.
// The controller reconciles helm charts available in the snap with the microcluster database until all valid charts are inserted.
func (c *HelmChartController) Run(ctx context.Context, s state.State) {
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithValues("controller", "helm-chart"))
	log := log.FromContext(ctx)

	log.Info("Waiting for node to be ready")
	// wait for microcluster node to be ready
	c.waitReady()

	log.Info("Starting helm chart controller")

	for {
		log.Info("Reconciling helm charts")

		var retryRequired bool

		for _, chartFS := range charts.Charts() {
			entries, err := chartFS.ReadDir("charts")
			if err != nil {
				log.Error(err, "failed to read helm charts directory")
				retryRequired = true
				continue
			}

			for _, e := range entries {
				if err := c.reconcile(ctx, s, chartFS, e); err != nil {
					log.Error(err, "failed to reconcile helm chart")
					retryRequired = true
				}
			}
		}

		if !retryRequired {
			log.Info("Reconcilation of helm charts complete")
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

func (c *HelmChartController) reconcile(ctx context.Context, s state.State, chartFS *embed.FS, e fs.DirEntry) error {
	log := log.FromContext(ctx).WithValues("entry", e.Name())

	if e.IsDir() {
		// TODO(berkayoz): Add support for directories by creating a tarball(.tgz) automatically
		// Directories should be converted manually via `helm package <directory>` command until then
		log.Info("entry is a directory, skipping reconciliation")
		return nil
	}

	chartBytes, err := chartFS.ReadFile(filepath.Join("charts", e.Name()))
	if err != nil {
		return fmt.Errorf("failed to read entry: %w", err)
	}

	chart, err := loader.LoadArchive(bytes.NewReader(chartBytes))
	if err != nil {
		log.Error(err, "failed to parse entry, skipping reconciliation")
		return nil
	}

	chartEntry := &types.HelmChartEntry{
		Name:     chart.Metadata.Name,
		Version:  chart.Metadata.Version,
		Contents: chartBytes,
	}

	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err := database.InsertHelmChart(ctx, tx, chartEntry)
		return err
	}); err != nil {
		return fmt.Errorf("failed to insert helm chart: %w", err)
	}

	log.WithValues("chart", chart.Metadata.Name, "version", chart.Metadata.Version).Info("helm chart reconciled successfully")

	return nil
}
