package utils

import (
	"context"
	"database/sql"
	"fmt"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database/clusterconfigs"
	"github.com/canonical/microcluster/state"
)

// ConvertBootstrapToClusterConfig extracts the cluster config parts from the BootstrapConfig
// and maps them to a ClusterConfig.
func ConvertBootstrapToClusterConfig(b *apiv1.BootstrapConfig) clusterconfigs.ClusterConfig {
	return clusterconfigs.ClusterConfig{
		Cluster: clusterconfigs.Cluster{
			CIDR: b.ClusterCIDR,
		},
	}
}

// GetClusterConfig is a convenience wrapper around the database call to get the cluster config.
func GetClusterConfig(ctx context.Context, state *state.State) (clusterconfigs.ClusterConfig, error) {
	var clusterConfig clusterconfigs.ClusterConfig
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		clusterConfig, err = clusterconfigs.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get cluster config from database: %w", err)
		}
		return nil
	}); err != nil {
		return clusterconfigs.ClusterConfig{}, fmt.Errorf("failed to perform cluster config transaction request: %w", err)
	}

	return clusterConfig, nil
}
