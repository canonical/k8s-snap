package app

import (
	"context"
	"crypto/rsa"
	"fmt"

	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	"github.com/canonical/microcluster/v2/state"
)

func (a *App) onStart(ctx context.Context, s state.State) error {
	// start a goroutine to mark the node as running
	go func() {
		if err := a.markNodeReady(ctx, s); err != nil {
			log.FromContext(ctx).Error(err, "Failed to mark node as ready")
		}
	}()

	// start helm chart controller
	if a.helmChartController != nil {
		go a.helmChartController.Run(ctx, s)
	}

	// start feature manifest controller
	if a.featureManifestController != nil {
		go a.featureManifestController.Run(ctx, s)
	}

	// start node config controller
	if a.nodeConfigController != nil {
		go a.nodeConfigController.Run(ctx, func(ctx context.Context) (*rsa.PublicKey, error) {
			cfg, err := databaseutil.GetClusterConfig(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("failed to load RSA key from configuration: %w", err)
			}
			keyPEM := cfg.Certificates.GetK8sdPublicKey()
			key, err := pkiutil.LoadRSAPublicKey(cfg.Certificates.GetK8sdPublicKey())
			if err != nil && keyPEM != "" {
				return nil, fmt.Errorf("failed to load RSA key: %w", err)
			}
			return key, nil
		})
	}

	// start node label controller
	if a.nodeLabelController != nil {
		go a.nodeLabelController.Run(ctx)
	}

	// start control plane config controller
	if a.controlPlaneConfigController != nil {
		go a.controlPlaneConfigController.Run(ctx, func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}

	// start update node config controller
	if a.updateNodeConfigController != nil {
		go a.updateNodeConfigController.Run(ctx, func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}

	// start feature controller
	if a.featureController != nil {
		go a.featureController.Run(
			ctx,
			s,
			a.NotifyUpdateNodeConfigController,
		)
	}

	// start csrsigning controller
	if a.csrsigningController != nil {
		go a.csrsigningController.Run(
			ctx,
			func(ctx context.Context) (types.ClusterConfig, error) {
				return databaseutil.GetClusterConfig(ctx, s)
			},
		)
	}

	return nil
}
