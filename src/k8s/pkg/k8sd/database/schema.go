package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/canonical/lxd/lxd/db/schema"
	"github.com/canonical/microcluster/v3/cluster"
)

var (
	// SchemaExtensions defines the schema updates for the database.
	// SchemaExtensions are apply only.
	// Note(ben): Never change the order or remove a migration as this would break the internal microcluster counter!
	SchemaExtensions = []schema.Update{
		schemaApplyMigration("kubernetes-auth-tokens", "000-create.sql"),
		schemaApplyMigration("cluster-configs", "000-create.sql"),
		schemaApplyMigration("worker-tokens", "000-create.sql"),
		schemaApplyMigration("feature-status", "000-feature-status.sql"),
		schemaApplyMigration("worker-tokens", "001-add-expiry.sql"),
	}

	//go:embed sql/migrations
	sqlMigrations embed.FS

	//go:embed sql/queries
	sqlQueries embed.FS
)

func schemaApplyMigration(migrationPath ...string) schema.Update {
	path := filepath.Join(append([]string{"sql", "migrations"}, migrationPath...)...)
	b, err := sqlMigrations.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("invalid migration file %s: %s", path, err))
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", path, err)
		}
		return nil
	}
}

// MustPrepareStatement reads and registers a SQL query.
func MustPrepareStatement(queryPath ...string) int {
	path := filepath.Join(append([]string{"sql", "queries"}, queryPath...)...)
	b, err := sqlQueries.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("invalid query file %s: %s", path, err))
	}
	return cluster.RegisterStmt(string(b))
}
