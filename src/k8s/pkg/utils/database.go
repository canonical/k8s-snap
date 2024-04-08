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

// CheckWorkerExists is a convenience wrapper around the database call to check if a worker node entry exists.
func CheckWorkerExists(ctx context.Context, state *state.State, name string) (bool, error) {
	var exists bool
	var err error

	if err := state.Database.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		exists, err = database.CheckWorkerExists(ctx, tx, name)
		if err != nil {
			return fmt.Errorf("failed to get worker node from database: %w", err)
		}
		return nil
	}); err != nil {
		return false, fmt.Errorf("failed to perform check worker node transaction request: %w", err)
	}

	return exists, nil
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
