package utils

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/state"
)

// GetClusterConfig is a convenience wrapper around the database call to get the cluster config.
func GetClusterConfig(ctx context.Context, state *state.State) (types.ClusterConfig, error) {
	var clusterConfig types.ClusterConfig
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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

// GetWorkerNodes is a convenience wrapper around the database call to get the worker node names.
func GetWorkerNodes(ctx context.Context, state *state.State) ([]string, error) {
	var workerNodes []string
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		workerNodes, err = database.ListWorkerNodes(ctx, tx)
		if err != nil {
			return fmt.Errorf("failed to list worker nodes from database: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to perform list worker nodes transaction request: %w", err)
	}

	return workerNodes, nil
}

// DeleteWorkerNodeEntry is a convenience wrapper around the database call to delete the worker node entry.
func DeleteWorkerNodeEntry(ctx context.Context, state *state.State, name string) error {
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		err = database.DeleteWorkerNode(ctx, tx, name)
		if err != nil {
			return fmt.Errorf("failed to delete worker node from database: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to perform delete worker node transaction request: %w", err)
	}

	return nil
}
