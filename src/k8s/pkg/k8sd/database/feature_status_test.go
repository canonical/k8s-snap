package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureStatus(t *testing.T) {
	WithDB(t, func(ctx context.Context, db DB) {
		_ = db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			t0, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			// initial get should return nothing
			ss, err := database.GetFeatureStatuses(ctx, tx)
			require.NoError(t, err)
			assert.Len(t, ss, 0)

			networkS := types.FeatureStatus{
				Enabled:   true,
				Message:   "enabled",
				Version:   "1.2.3",
				UpdatedAt: t0,
			}
			dnsS := types.FeatureStatus{
				Enabled:   true,
				Message:   "enabled at 10.0.0.1",
				Version:   "4.5.6",
				UpdatedAt: t0,
			}
			// setting new values
			err = database.SetFeatureStatus(ctx, tx, "network", networkS)
			require.NoError(t, err)
			err = database.SetFeatureStatus(ctx, tx, "dns", dnsS)
			require.NoError(t, err)

			// getting new values
			ss, err = database.GetFeatureStatuses(ctx, tx)
			require.NoError(t, err)
			assert.Len(t, ss, 2)

			assert.Equal(t, networkS.Enabled, ss["network"].Enabled)
			assert.Equal(t, networkS.Message, ss["network"].Message)
			assert.Equal(t, networkS.Version, ss["network"].Version)
			assert.Equal(t, networkS.UpdatedAt, ss["network"].UpdatedAt)

			assert.Equal(t, dnsS.Enabled, ss["dns"].Enabled)
			assert.Equal(t, dnsS.Message, ss["dns"].Message)
			assert.Equal(t, dnsS.Version, ss["dns"].Version)
			assert.Equal(t, dnsS.UpdatedAt, ss["dns"].UpdatedAt)

			// updating old values and adding new ones
			dnsS2 := types.FeatureStatus{
				Enabled:   true,
				Message:   "enabled at 10.0.0.2",
				Version:   "4.5.7",
				UpdatedAt: t0,
			}
			gatewayS := types.FeatureStatus{
				Enabled:   true,
				Message:   "disabled",
				Version:   "10.20.30",
				UpdatedAt: t0,
			}
			// setting the old value for network again
			err = database.SetFeatureStatus(ctx, tx, "network", networkS)
			require.NoError(t, err)
			// updating dns with new value
			err = database.SetFeatureStatus(ctx, tx, "dns", dnsS2)
			require.NoError(t, err)
			// adding new status
			err = database.SetFeatureStatus(ctx, tx, "gateway", gatewayS)
			require.NoError(t, err)

			// checking the new values
			ss, err = database.GetFeatureStatuses(ctx, tx)
			require.NoError(t, err)
			assert.Len(t, ss, 3)

			// network stayed the same
			assert.Equal(t, networkS.Enabled, ss["network"].Enabled)
			assert.Equal(t, networkS.Message, ss["network"].Message)
			assert.Equal(t, networkS.Version, ss["network"].Version)
			assert.Equal(t, networkS.UpdatedAt, ss["network"].UpdatedAt)

			// dns is updated
			assert.Equal(t, dnsS2.Enabled, ss["dns"].Enabled)
			assert.Equal(t, dnsS2.Message, ss["dns"].Message)
			assert.Equal(t, dnsS2.Version, ss["dns"].Version)
			assert.Equal(t, dnsS2.UpdatedAt, ss["dns"].UpdatedAt)

			// gateway is added
			assert.Equal(t, gatewayS.Enabled, ss["gateway"].Enabled)
			assert.Equal(t, gatewayS.Message, ss["gateway"].Message)
			assert.Equal(t, gatewayS.Version, ss["gateway"].Version)
			assert.Equal(t, gatewayS.UpdatedAt, ss["gateway"].UpdatedAt)

			return nil
		})
	})
}
