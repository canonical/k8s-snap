package databaseutil

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/microcluster/state"
)

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

// IsWorkerNode returns true if the given node name belongs to a worker node in the cluster.
func IsWorkerNode(ctx context.Context, s *state.State, name string) (bool, error) {
	exists, err := CheckWorkerExists(ctx, s, name)
	if err != nil {
		return false, fmt.Errorf("failed to check if worker node %q exists: %w", name, err)
	}
	return exists, nil
}
