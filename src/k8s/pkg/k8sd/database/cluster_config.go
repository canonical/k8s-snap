package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v3/cluster"
)

var clusterConfigsStmts = map[string]int{
	"insert-v1alpha2": MustPrepareStatement("cluster-configs", "insert-v1alpha2.sql"),
	"select-v1alpha2": MustPrepareStatement("cluster-configs", "select-v1alpha2.sql"),
}

// SetClusterConfig updates the cluster configuration with any non-empty values that are set.
// SetClusterConfig will attempt to merge the existing and new configs, and return an error if any protected fields have changed.
// SetClusterConfig will return the merged cluster configuration on success.
func SetClusterConfig(ctx context.Context, tx *sql.Tx, new types.ClusterConfig) (types.ClusterConfig, error) {
	old, err := GetClusterConfig(ctx, tx)
	if err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to fetch existing cluster config: %w", err)
	}
	config, err := types.MergeClusterConfig(old, new)
	if err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to merge new cluster configuration options: %w", err)
	}

	b, err := json.Marshal(config)
	if err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to encode cluster config: %w", err)
	}
	insertTxStmt, err := cluster.Stmt(tx, clusterConfigsStmts["insert-v1alpha2"])
	if err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, string(b)); err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to insert v1alpha2 config: %w", err)
	}
	return config, nil
}

// GetClusterConfig retrieves the cluster configuration from the database.
func GetClusterConfig(ctx context.Context, tx *sql.Tx) (types.ClusterConfig, error) {
	txStmt, err := cluster.Stmt(tx, clusterConfigsStmts["select-v1alpha2"])
	if err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	var s string
	if err := txStmt.QueryRowContext(ctx).Scan(&s); err != nil {
		if err == sql.ErrNoRows {
			return types.ClusterConfig{}, nil
		}
		return types.ClusterConfig{}, fmt.Errorf("failed to retrieve v1alpha2 config: %w", err)
	}

	var clusterConfig types.ClusterConfig
	if err := json.Unmarshal([]byte(s), &clusterConfig); err != nil {
		return types.ClusterConfig{}, fmt.Errorf("failed to parse v1alpha2 config: %w", err)
	}

	return clusterConfig, nil
}
