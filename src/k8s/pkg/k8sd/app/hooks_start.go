package app

import (
	"context"
	"crypto/rsa"
	"fmt"

	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/state"
)

func (a *App) onStart(s *state.State) error {
	// start a goroutine to mark the node as running
	go a.markNodeReady(s.Context, s)

	// start node config controller
	if a.nodeConfigController != nil {
		go a.nodeConfigController.Run(s.Context, func(ctx context.Context) (*rsa.PublicKey, error) {
			cfg, err := databaseutil.GetClusterConfig(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("failed to load RSA key from configuration: %w", err)
			}
			keyPEM := cfg.Certificates.GetK8sdPublicKey()
			key, err := pki.LoadRSAPublicKey(cfg.Certificates.GetK8sdPublicKey())
			if err != nil && keyPEM != "" {
				return nil, fmt.Errorf("failed to load RSA key: %w", err)
			}
			return key, nil
		})
	}

	// start control plane config controller
	if a.controlPlaneConfigController != nil {
		go a.controlPlaneConfigController.Run(s.Context, func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}

	// start update node config controller
	if a.updateNodeConfigController != nil {
		go a.updateNodeConfigController.Run(s.Context, func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}

	return nil
}
