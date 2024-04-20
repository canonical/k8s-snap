package app

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/pki"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/state"
)

func (a *App) onStart(s *state.State) error {
	// start a goroutine to mark the node as running
	go a.markNodeReady(s.Context, s)

	// start node config controller
	if a.nodeConfigController != nil {
		go a.nodeConfigController.Run(s.Context, func(ctx context.Context) (*rsa.PublicKey, error) {
			cfg, err := utils.GetClusterConfig(ctx, s)
			if err != nil {
				return nil, fmt.Errorf("failed to load RSA key from configuration: %w", err)
			}
			key, err := pki.LoadRSAPublicKey(cfg.Certificates.GetK8sdPublicKey())
			if err != nil {
				return nil, fmt.Errorf("failed to load RSA key: %w", err)
			}
			return key, nil
		})
	}

	// start control plane config controller
	if a.controlPlaneConfigController != nil {
		go a.controlPlaneConfigController.Run(s.Context, func(ctx context.Context) (types.ClusterConfig, error) {
			return utils.GetClusterConfig(ctx, s)
		})
	}

	// start update node config controller
	if a.updateNodeConfigController != nil {
		go a.updateNodeConfigController.Run(s.Context, func(ctx context.Context) (types.ClusterConfig, error) {
			return utils.GetClusterConfig(ctx, s)
		})
	}

	return nil
}
