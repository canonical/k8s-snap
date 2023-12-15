package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/canonical/microcluster/cluster"
)

var (
	certificatesStmts = map[string]int{
		"insert-certificate": mustPrepareStatement("certificates", "insert-certificate.sql"),
		"select-by-name":     mustPrepareStatement("certificates", "select-by-name.sql"),
	}
)

// CreateCertificate inserts a new certificate entry.
// TODO: convert this to Upsert if necessary.
func CreateCertificate(ctx context.Context, tx *sql.Tx, name string, certificate string, key string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if certificate == "" {
		return fmt.Errorf("certificate cannot be empty")
	}
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	insertTxStmt, err := cluster.Stmt(tx, certificatesStmts["insert-certificate"])
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, name, key, certificate); err != nil {
		return fmt.Errorf("insert certificate query failed: %w", err)
	}

	return nil
}

// GetCertificateAndKey retrieves the certificate and key for the given name.
func GetCertificateAndKey(ctx context.Context, tx *sql.Tx, name string) (cert string, key string, err error) {
	if name == "" {
		return "", "", fmt.Errorf("name cannot be empty")
	}
	txStmt, err := cluster.Stmt(tx, certificatesStmts["select-by-name"])
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare statement: %w", err)
	}
	if err := txStmt.QueryRowContext(ctx, name).Scan(&cert, &key); err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("no cert for name=%s: %w", name, err)
		}
		return "", "", fmt.Errorf("failed to get certificate: %w", err)
	}

	return cert, key, nil
}
