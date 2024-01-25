package database

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/canonical/microcluster/cluster"
)

var (
	workerStmts = map[string]int{
		"insert-node": mustPrepareStatement("worker-nodes", "insert.sql"),
		"select-node": mustPrepareStatement("worker-nodes", "select.sql"),
		"delete-node": mustPrepareStatement("worker-nodes", "delete.sql"),

		"insert-token": mustPrepareStatement("cluster-configs", "insert-worker-token.sql"),
		"select-token": mustPrepareStatement("cluster-configs", "select-worker-token.sql"),
		"delete-token": mustPrepareStatement("cluster-configs", "delete-worker-token.sql"),
	}
)

// CheckWorkerNodeToken returns true if the specified can be used to join worker nodes on the cluster.
func CheckWorkerNodeToken(ctx context.Context, tx *sql.Tx, token string) (bool, error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-token"])
	if err != nil {
		return false, fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var realToken string
	if selectTxStmt.QueryRowContext(ctx).Scan(&realToken) == nil {
		return subtle.ConstantTimeCompare([]byte(token), []byte(realToken)) == 1, nil
	}
	return false, nil
}

// GetOrCreateWorkerNodeToken returns a token that can be used to join worker nodes on the cluster.
func GetOrCreateWorkerNodeToken(ctx context.Context, tx *sql.Tx) (string, error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var token string
	if selectTxStmt.QueryRowContext(ctx).Scan(&token) == nil {
		return token, nil
	}

	// generate random bytes for the token
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("is the system entropy low? failed to get random bytes: %w", err)
	}
	token = fmt.Sprintf("worker::%s", hex.EncodeToString(b))

	insertTxStmt, err := cluster.Stmt(tx, workerStmts["insert-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, token); err != nil {
		return "", fmt.Errorf("insert token query failed: %w", err)
	}
	return token, nil
}

// DeleteWorkerNodeToken returns a token that can be used to join worker nodes on the cluster.
func DeleteWorkerNodeToken(ctx context.Context, tx *sql.Tx) error {
	deleteTxStmt, err := cluster.Stmt(tx, workerStmts["delete-token"])
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	if _, err := deleteTxStmt.ExecContext(ctx); err != nil {
		return fmt.Errorf("delete token query failed: %w", err)
	}
	return nil
}

// AddWorkerNode adds a new worker node entry on the database.
func AddWorkerNode(ctx context.Context, tx *sql.Tx, name string) error {
	insertTxStmt, err := cluster.Stmt(tx, workerStmts["insert-node"])
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, name); err != nil {
		return fmt.Errorf("insert worker node query failed: %w", err)
	}
	return nil
}

// ListWorkerNodes lists the known worker nodes on the database.
func ListWorkerNodes(ctx context.Context, tx *sql.Tx) ([]string, error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-node"])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}
	rows, err := selectTxStmt.QueryContext(ctx)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("select worker nodes query failed: %w", err)
	}
	var nodes []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to parse row: %w", err)
		}
		nodes = append(nodes, name)
	}

	return nodes, nil
}

// DeleteWorkerNode deletes a new worker node from the database.
func DeleteWorkerNode(ctx context.Context, tx *sql.Tx, name string) error {
	deleteTxStmt, err := cluster.Stmt(tx, workerStmts["delete-node"])
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	if _, err := deleteTxStmt.ExecContext(ctx, name); err != nil {
		return fmt.Errorf("insert worker node query failed: %w", err)
	}
	return nil
}
