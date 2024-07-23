package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/cluster"
)

var featureStatusStmts = struct {
	select_ int
	upsert_ int
}{
	select_: MustPrepareStatement("feature-status", "select.sql"),
	upsert_: MustPrepareStatement("feature-status", "upsert.sql"),
}

// SetFeatureStatus updates the status of the given feature.
func SetFeatureStatus(ctx context.Context, tx *sql.Tx, name string, status types.FeatureStatus) error {
	upsertTxStmt, err := cluster.Stmt(tx, featureStatusStmts.upsert_)
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}

	if _, err := upsertTxStmt.ExecContext(ctx,
		name,
		status.Message,
		status.Version,
		status.UpdatedAt.Format(time.RFC3339),
		status.Enabled,
	); err != nil {
		return fmt.Errorf("failed to execute upsert statement: %w", err)
	}

	return nil
}

// GetFeatureStatuses returns a map of feature names to their status.
func GetFeatureStatuses(ctx context.Context, tx *sql.Tx) (map[string]types.FeatureStatus, error) {
	selectTxStmt, err := cluster.Stmt(tx, featureStatusStmts.select_)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	rows, err := selectTxStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select statement: %w", err)
	}

	fsMap := make(map[string]types.FeatureStatus)

	for rows.Next() {
		var (
			name string
			ts   string
		)
		typ := types.FeatureStatus{}

		if err := rows.Scan(&name, &typ.Message, &typ.Version, &ts, &typ.Enabled); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		typ.UpdatedAt, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp: %w", err)
		}

		fsMap[name] = typ
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	return fsMap, nil
}
