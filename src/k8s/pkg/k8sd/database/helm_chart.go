package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v2/cluster"
)

var chartStmts = map[string]int{
	"insert-chart": MustPrepareStatement("helm-charts", "insert.sql"),
	"select-all":   MustPrepareStatement("helm-charts", "select.sql"),
	"select-chart": MustPrepareStatement("helm-charts", "select-chart.sql"),
}

// GetFeatureStatuses returns a map of feature names to their status.
func GetHelmCharts(ctx context.Context, tx *sql.Tx) ([]types.HelmChart, error) {
	selectTxStmt, err := cluster.Stmt(tx, chartStmts["select-all"])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	rows, err := selectTxStmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select statement: %w", err)
	}

	var result []types.HelmChart

	for rows.Next() {
		chart := types.HelmChart{}

		if err := rows.Scan(&chart.Name, &chart.Version, &chart.Contents); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		result = append(result, chart)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	return result, nil
}

// InsertHelmChart inserts a helm chart into the database.
func InsertHelmChart(ctx context.Context, tx *sql.Tx, name string, version string, contents []byte) error {
	insertTxStmt, err := cluster.Stmt(tx, chartStmts["insert-chart"])
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}

	if _, err := insertTxStmt.ExecContext(ctx,
		name,
		version,
		contents,
	); err != nil {
		return fmt.Errorf("failed to execute upsert statement: %w", err)
	}

	return nil
}
