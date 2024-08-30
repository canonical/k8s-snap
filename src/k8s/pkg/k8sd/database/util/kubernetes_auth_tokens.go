package databaseutil

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/microcluster/v3/state"
)

// GetOrCreateAuthToken returns a k8s auth token based on the provided username/groups.
func GetOrCreateAuthToken(ctx context.Context, state state.State, username string, groups []string) (string, error) {
	var token string
	if err := state.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetOrCreateToken(ctx, tx, username, groups)
		return err
	}); err != nil {
		return "", fmt.Errorf("database transaction failed: %w", err)
	}
	return token, nil
}

func RevokeAuthToken(ctx context.Context, state state.State, token string) error {
	if err := state.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if err := database.DeleteToken(ctx, tx, token); err != nil {
			return fmt.Errorf("failed to delete token from database: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction failed: %w", err)
	}
	return nil
}
