package database

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/canonical/microcluster/v3/cluster"
)

var workerStmts = map[string]int{
	"insert-token": MustPrepareStatement("worker-tokens", "insert.sql"),
	"select-token": MustPrepareStatement("worker-tokens", "select.sql"),
	"delete-token": MustPrepareStatement("worker-tokens", "delete-by-token.sql"),
}

// CheckWorkerNodeToken returns true if the specified token can be used to join the specified node on the cluster.
// CheckWorkerNodeToken will return true if the token is empty or if the token is associated with the specified node
// and has not expired.
func CheckWorkerNodeToken(ctx context.Context, tx *sql.Tx, nodeName string, token string) (bool, error) {
	selectTxStmt, err := cluster.Stmt(tx, workerStmts["select-token"])
	if err != nil {
		return false, fmt.Errorf("failed to prepare select statement: %w", err)
	}
	var tokenNodeName string
	var expiry time.Time
	if selectTxStmt.QueryRowContext(ctx, token).Scan(&tokenNodeName, &expiry) == nil {
		isValidToken := tokenNodeName == "" || subtle.ConstantTimeCompare([]byte(nodeName), []byte(tokenNodeName)) == 1
		notExpired := time.Now().Before(expiry)
		return isValidToken && notExpired, nil
	}
	return false, nil
}

// GetOrCreateWorkerNodeToken returns a token that can be used to join a worker node on the cluster.
// GetOrCreateWorkerNodeToken will return the existing token, if one already exists for the node.
func GetOrCreateWorkerNodeToken(ctx context.Context, tx *sql.Tx, nodeName string, expiry time.Time) (string, error) {
	insertTxStmt, err := cluster.Stmt(tx, workerStmts["insert-token"])
	if err != nil {
		return "", fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	// generate random bytes for the token
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("is the system entropy low? failed to get random bytes: %w", err)
	}
	token := fmt.Sprintf("worker::%s", hex.EncodeToString(b))
	if _, err := insertTxStmt.ExecContext(ctx, nodeName, token, expiry); err != nil {
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
	if _, err := deleteTxStmt.ExecContext(ctx, nodeName); err != nil {
		return fmt.Errorf("delete token query failed: %w", err)
	}
	return nil
}
