package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/canonical/lxd/lxd/db/schema"
)

var (
	SchemaExtensions = map[int]schema.Update{
		1: schemaApplyMigration("000-k8sd-tokens-create.sql"),
	}

	//go:embed sql/migrations
	sqlMigrations embed.FS
)

func schemaApplyMigration(migrationName string) schema.Update {
	b, err := sqlMigrations.ReadFile(filepath.Join("sql", "migrations", migrationName))
	if err != nil {
		panic(fmt.Errorf("migration %q not defined: %s", migrationName, err))
	}
	return func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, string(b)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migrationName, err)
		}
		return nil
	}
}
