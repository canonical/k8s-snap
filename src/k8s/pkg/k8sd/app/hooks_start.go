package app

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
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
		go a.controlPlaneConfigController.Run(
			s.Context,
			func(ctx context.Context) (types.ClusterConfig, []string, error) {
				cfg, err := databaseutil.GetClusterConfig(ctx, s)
				if err != nil {
					return types.ClusterConfig{}, nil, fmt.Errorf("failed to retrieve cluster config: %w", err)
				}

				if cfg.Datastore.GetType() != "etcd" {
					return cfg, nil, nil
				}
				leader, err := s.Leader()
				if err != nil {
					return types.ClusterConfig{}, nil, fmt.Errorf("failed to retrieve cluster leader: %w", err)
				}
				members, err := leader.GetClusterMembers(ctx)
				if err != nil {
					return types.ClusterConfig{}, nil, fmt.Errorf("failed to retrieve cluster members: %w", err)
				}
				nodeIPs := make([]string, 0, len(members))
				for _, member := range members {
					nodeIPs = append(nodeIPs, member.Address.Addr().String())
				}

				return cfg, nodeIPs, nil
			},
		)
	}

	// start update node config controller
	if a.updateNodeConfigController != nil {
		go a.updateNodeConfigController.Run(s.Context, func(ctx context.Context) (types.ClusterConfig, error) {
			return databaseutil.GetClusterConfig(ctx, s)
		})
	}

	// start feature controller
	if a.featureController != nil {
		go a.featureController.Run(
			s.Context,
			func(ctx context.Context) (types.ClusterConfig, error) {
				return databaseutil.GetClusterConfig(ctx, s)
			},
			func(ctx context.Context, dnsIP string) error {
				if err := s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
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

	return nil
}
