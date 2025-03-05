package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v2/cluster"
	"gopkg.in/yaml.v2"
)

var featureStmts = map[string]int{
	"insert": MustPrepareStatement("feature-manifests", "insert.sql"),
	"select": MustPrepareStatement("feature-manifests", "select.sql"),
}

// GetFeatureManifest returns a feature manifest from the database by name and version.
func GetFeatureManifest(ctx context.Context, tx *sql.Tx, name string, version string) (*types.FeatureManifest, error) {
	selectTxStmt, err := cluster.Stmt(tx, featureStmts["select"])
	if err != nil {
		return nil, fmt.Errorf("failed to prepare select statement: %w", err)
	}

	var manifestBytes []byte

	err = selectTxStmt.QueryRowContext(ctx, name, version).Scan(&name, &version, &manifestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to query helm chart: %w", err)
	}

	var manifest types.FeatureManifest

	if err := yaml.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	return &manifest, nil
}

// InsertFeatureManifest inserts a feature manifest into the database.
func InsertFeatureManifest(ctx context.Context, tx *sql.Tx, manifest *types.FeatureManifest) error {
	insertTxStmt, err := cluster.Stmt(tx, featureStmts["insert"])
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}

	manifestBytes, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if _, err := insertTxStmt.ExecContext(ctx,
		manifest.Name,
		manifest.Version,
		manifestBytes,
	); err != nil {
		return fmt.Errorf("failed to execute upsert statement: %w", err)
	}

	return nil
}
