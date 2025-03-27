package app

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/utils"
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

	// Check if a refresh was performed and if so, run the custom post-refresh hook
	isPostRefresh, err := utils.FileExists(a.snap.PostRefreshLockPath())
	if err != nil {
		return fmt.Errorf("failed to check if snap is post-refresh: %w", err)
	}
	if isPostRefresh {
		if err := a.postRefreshHook(ctx, s); err != nil {
			return fmt.Errorf("failed to run post-refresh hook: %w", err)
		}
		if err := os.Remove(a.snap.PostRefreshLockPath()); err != nil {
			return fmt.Errorf("failed to remove post-refresh lock file: %w", err)
		}
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
			func(ctx context.Context) (types.ClusterConfig, error) {
				return databaseutil.GetClusterConfig(ctx, s)
			},
			func(ctx context.Context, dnsIP string) error {
				if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					if _, err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
						Kubelet: types.Kubelet{ClusterDNS: utils.Pointer(dnsIP)},
					}); err != nil {
						return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
					}
					return nil
				}); err != nil {
					return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
				}

				// DNS IP has changed, notify node config controller
				a.NotifyUpdateNodeConfigController()

				return nil
			},
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
