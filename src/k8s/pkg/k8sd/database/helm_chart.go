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
	"select":       MustPrepareStatement("helm-charts", "select.sql"),
}

// GetHelmChart returns a helm chart from the database by name and version.
func GetHelmChart(ctx context.Context, tx *sql.Tx, name string, version string) (*types.HelmChartEntry, error) {
	selectTxStmt, err := cluster.Stmt(tx, chartStmts["select"])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	chart := types.HelmChartEntry{}

	err = selectTxStmt.QueryRowContext(ctx, name, version).Scan(&chart.Name, &chart.Version, &chart.Contents)
	if err != nil {
		return nil, fmt.Errorf("failed to query helm chart: %w", err)
	}

	return &chart, nil
}

// InsertHelmChart inserts a helm chart into the database.
func InsertHelmChart(ctx context.Context, tx *sql.Tx, chartEntry *types.HelmChartEntry) error {
	insertTxStmt, err := cluster.Stmt(tx, chartStmts["insert-chart"])
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}

	if _, err := insertTxStmt.ExecContext(ctx,
		chartEntry.Name,
		chartEntry.Version,
		chartEntry.Contents,
	); err != nil {
		return fmt.Errorf("failed to execute upsert statement: %w", err)
	}

	return nil
}
