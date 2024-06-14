package database

import (
	"context"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/canonical/microcluster/cluster"
)

var (
	capiTokenStmts = map[string]int{
		"insert-token": MustPrepareStatement("capi-auth", "insert-token.sql"),
		"select-token": MustPrepareStatement("capi-auth", "select.sql"),
	}
)

// CheckAuthToken returns true if the specified token matches the ClusterAPI token.
func CheckAuthToken(ctx context.Context, tx *sql.Tx, token string) (bool, error) {
	selectTxStmt, err := cluster.Stmt(tx, capiTokenStmts["select-token"])
	if err != nil {
		return false, fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var tokenID int32
	if selectTxStmt.QueryRowContext(ctx, token).Scan(&tokenID) == nil {
		return subtle.ConstantTimeEq(tokenID, 1) == 1, nil
	}
	return false, nil
}

// SetAuthToken sets the ClusterAPI token.
func SetAuthToken(ctx context.Context, tx *sql.Tx, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	insertTxStmt, err := cluster.Stmt(tx, capiTokenStmts["insert-token"])
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, token); err != nil {
		return fmt.Errorf("insert token query failed: %w", err)
	}

	return nil
}
