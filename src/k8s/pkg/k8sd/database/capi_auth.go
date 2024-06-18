package database

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
)

// SetClusterAPIToken stores the ClusterAPI token in the cluster config.
func SetClusterAPIToken(ctx context.Context, tx *sql.Tx, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	cfg := types.ClusterConfig{
		ClusterAPI: types.ClusterAPI{
			AuthToken: utils.Pointer(token),
		},
	}
	if _, err := SetClusterConfig(ctx, tx, cfg); err != nil {
		return fmt.Errorf("failed to write cluster configuration: %w", err)
	}
	return nil

}

// ValidateClusterAPIToken returns true if the specified token matches the stored ClusterAPI token.
func ValidateClusterAPIToken(ctx context.Context, tx *sql.Tx, token string) (bool, error) {
	cfg, err := GetClusterConfig(ctx, tx)
	if err != nil {
		return false, fmt.Errorf("failed to fetch existing ClusterAPI token: %w", err)
	}
	if cfg.ClusterAPI.AuthToken != nil {
		return subtle.ConstantTimeCompare([]byte(token), []byte(*cfg.ClusterAPI.AuthToken)) == 1, nil
	}
	return false, nil
}
