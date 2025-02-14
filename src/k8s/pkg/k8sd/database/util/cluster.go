package databaseutil

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v2/state"
)

// GetClusterConfig is a convenience wrapper around the database call to get the cluster config.
func GetClusterConfig(ctx context.Context, state state.State) (types.ClusterConfig, error) {
	var clusterConfig types.ClusterConfig
	var err error

	if err := state.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		clusterConfig, err = database.GetClusterConfig(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to get cluster config from database: %w", err)
		}
		return nil
	}); err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to perform cluster config transaction request: %w", err)
	}

	return clusterConfig, nil
}

// GetClusterBootstrapConfig is a convenience wrapper around the database call to get the cluster bootstrap config.
func GetClusterBootstrapConfig(ctx context.Context, state state.State) (types.ClusterConfig, error) {
	var config types.ClusterConfig

	if err := state.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		if config, err = database.GetClusterBootstrapConfig(ctx, tx); err != nil {
			return fmt.Errorf("failed to get cluster bootstrap config from database: %w", err)
		}
		return nil
	}); err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to perform get cluster bootstrap config transaction: %w", err)
	}

	return config, nil
}
