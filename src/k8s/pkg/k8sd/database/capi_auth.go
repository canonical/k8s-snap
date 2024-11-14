package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/microcluster/v2/cluster"
)

var clusterAPIConfigsStmts = map[string]int{
	"insert-capi-token": MustPrepareStatement("cluster-configs", "insert-capi-token.sql"),
	"select-capi-token": MustPrepareStatement("cluster-configs", "select-capi-token.sql"),
}

// SetClusterAPIToken stores the ClusterAPI token in the cluster config.
func SetClusterAPIToken(ctx context.Context, tx *sql.Tx, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	insertTxStmt, err := cluster.Stmt(tx, clusterAPIConfigsStmts["insert-capi-token"])
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, token); err != nil {
		return fmt.Errorf("insert ClusterAPI token query failed: %w", err)
	}

	return nil
}

// ValidateClusterAPIToken returns true if the specified token matches the stored ClusterAPI token.
func ValidateClusterAPIToken(ctx context.Context, tx *sql.Tx, token string) (bool, error) {
	selectTxStmt, err := cluster.Stmt(tx, clusterAPIConfigsStmts["select-capi-token"])
	if err != nil {
		return false, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	var exists bool
	err = selectTxStmt.QueryRowContext(ctx, token).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to query ClusterAPI token: %w", err)
	}

	return exists, nil
}
