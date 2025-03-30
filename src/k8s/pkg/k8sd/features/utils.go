package features

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
)

func UpdateClusterDNS(ctx context.Context, s state.State, dnsIP string) error {
	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
			Kubelet: types.Kubelet{ClusterDNS: utils.Pointer(dnsIP)},
		}); err != nil {
			return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
	}

	return nil
}
