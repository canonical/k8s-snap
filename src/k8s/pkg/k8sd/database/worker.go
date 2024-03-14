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
		"insert-node":    MustPrepareStatement("worker-nodes", "insert.sql"),
		"select-node":    MustPrepareStatement("worker-nodes", "select.sql"),
		"select-by-name": MustPrepareStatement("worker-nodes", "select-by-name.sql"),
		"delete-node":    MustPrepareStatement("worker-nodes", "delete.sql"),

		"insert-token": MustPrepareStatement("cluster-configs", "insert-worker-token.sql"),
		"select-token": MustPrepareStatement("cluster-configs", "select-worker-token.sql"),
		"delete-token": MustPrepareStatement("cluster-configs", "delete-worker-token.sql"),
	}
)

// CheckWorkerNodeToken returns true if the specified token can be used to join the specified node on the cluster.
func CheckWorkerNodeToken(ctx context.Context, tx *sql.Tx, nodeName string, token string) (bool, error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-token"])
	if err != nil {
		return false, fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var realToken string
	if selectTxStmt.QueryRowContext(ctx, fmt.Sprintf("worker-token::%s", nodeName)).Scan(&realToken) == nil {
		return subtle.ConstantTimeCompare([]byte(token), []byte(realToken)) == 1, nil
	}
	return false, nil
}

// GetOrCreateWorkerNodeToken returns a token that can be used to join a worker node on the cluster.
// GetOrCreateWorkerNodeToken will return the existing token, if one already exists for the node.
func GetOrCreateWorkerNodeToken(ctx context.Context, tx *sql.Tx, nodeName string) (string, error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var token string
	if selectTxStmt.QueryRowContext(ctx, fmt.Sprintf("worker-token::%s", nodeName)).Scan(&token) == nil {
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
	if _, err := insertTxStmt.ExecContext(ctx, fmt.Sprintf("worker-token::%s", nodeName), token); err != nil {
		return "", fmt.Errorf("insert token query failed: %w", err)
	}
	return token, nil
}

// DeleteWorkerNodeToken returns a token that can be used to join worker nodes on the cluster.
func DeleteWorkerNodeToken(ctx context.Context, tx *sql.Tx, nodeName string) error {
	deleteTxStmt, err := cluster.Stmt(tx, workerStmts["delete-token"])
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	if _, err := deleteTxStmt.ExecContext(ctx, fmt.Sprintf("worker-token::%s", nodeName)); err != nil {
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

// CheckWorkerExists returns true if a worker node entry for this name exists.
func CheckWorkerExists(ctx context.Context, tx *sql.Tx, name string) (exists bool, err error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-by-name"])
	if err != nil {
		return false, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	row := selectTxStmt.QueryRowContext(ctx, name)
	if err := row.Scan(new(string)); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("select worker node %q query failed: %w", name, err)
	}

	return true, nil
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

// DeleteWorkerNode deletes a worker node from the database.
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
