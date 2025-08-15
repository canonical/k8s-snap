package app

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"time"

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

	// tune system settings if enabled and necessary
	if !a.snap.Strict() {
		if err := a.tuneSystemSettings(ctx, s); err != nil {
			log.FromContext(ctx).Error(err, "failed to tune system settings")
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
		go a.nodeLabelController.Run(ctx, func(ctx context.Context) (string, error) {
			cfg, err := databaseutil.GetClusterConfig(ctx, s)
			if err != nil {
				return "", fmt.Errorf("failed to retrieve cluster config: %w", err)
			}

			datastoreType := cfg.Datastore.Type
			if datastoreType == nil {
				return "", fmt.Errorf("datastore type is not set in cluster config")
			}

			return *datastoreType, nil
		})
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
			func(ctx context.Context) (types.ClusterConfig, error) {
				return databaseutil.GetClusterConfig(ctx, s)
			},
			func() state.State {
				return s
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
			func(ctx context.Context, name types.FeatureName, featureStatus types.FeatureStatus) error {
				if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					// we set timestamp here in order to reduce the clutter. otherwise we will need to
					// set .UpdatedAt field in a lot of places for every event/error.
					// this is not 100% accurate but should be good enough
					featureStatus.UpdatedAt = time.Now()
					if err := database.SetFeatureStatus(ctx, tx, name, featureStatus); err != nil {
						return fmt.Errorf("failed to set feature status in db for %q: %w", name, err)
					}
					return nil
				}); err != nil {
					return fmt.Errorf("database transaction to set feature status failed: %w", err)
				}
				return nil
			},
		)
	}

	// start controller coordinator
	if a.controllerCoordinator != nil {
		go func() {
			if err := a.controllerCoordinator.Run(
				ctx,
				func(ctx context.Context) (types.ClusterConfig, error) {
					return databaseutil.GetClusterConfig(ctx, s)
				},
			); err != nil {
				log.FromContext(ctx).Error(err, "Failed to start controller coordinator")
			}
		}()
	}

	return nil
}
