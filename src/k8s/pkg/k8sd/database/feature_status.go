package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/microcluster/cluster"
)

var featureStatusStmts = map[string]int{
	"select": MustPrepareStatement("feature-status", "select.sql"),
	"upsert": MustPrepareStatement("feature-status", "upsert.sql"),
}

// SetFeatureStatus updates the status of the given feature.
func SetFeatureStatus(ctx context.Context, tx *sql.Tx, name string, status types.FeatureStatus) error {
	upsertTxStmt, err := cluster.Stmt(tx, featureStatusStmts["upsert"])
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
	selectTxStmt, err := cluster.Stmt(tx, featureStatusStmts["select"])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	rows, err := selectTxStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select statement: %w", err)
	}

	result := make(map[string]types.FeatureStatus)

	for rows.Next() {
		var (
			name string
			ts   string
		)
		status := types.FeatureStatus{}

		if err := rows.Scan(&name, &status.Message, &status.Version, &ts, &status.Enabled); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		status.UpdatedAt, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			log.L().Error(err, "failed to parse time", "original", ts)
		}

		result[name] = status
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	return result, nil
}
