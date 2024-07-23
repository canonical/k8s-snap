package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

func TestFeatureStatus(t *testing.T) {
	WithDB(t, func(ctx context.Context, db DB) {
		_ = db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			t0, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
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

			t.Run("ReturnNothingInitially", func(t *testing.T) {
				g := NewWithT(t)
				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(len(ss)).To(Equal(0))

			})

			t.Run("SettingNewStatus", func(t *testing.T) {
				g := NewWithT(t)

				err := database.SetFeatureStatus(ctx, tx, "network", networkS)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "dns", dnsS)
				g.Expect(err).To(BeNil())

				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(len(ss)).To(Equal(2))

				g.Expect(ss["network"].Enabled).To(Equal(networkS.Enabled))
				g.Expect(ss["network"].Message).To(Equal(networkS.Message))
				g.Expect(ss["network"].Version).To(Equal(networkS.Version))
				g.Expect(ss["network"].UpdatedAt).To(Equal(networkS.UpdatedAt))

				g.Expect(ss["dns"].Enabled).To(Equal(dnsS.Enabled))
				g.Expect(ss["dns"].Message).To(Equal(dnsS.Message))
				g.Expect(ss["dns"].Version).To(Equal(dnsS.Version))
				g.Expect(ss["dns"].UpdatedAt).To(Equal(dnsS.UpdatedAt))

			})
			t.Run("UpdatingStatus", func(t *testing.T) {
				g := NewWithT(t)

				err := database.SetFeatureStatus(ctx, tx, "network", networkS)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "dns", dnsS)
				g.Expect(err).To(BeNil())

				// set and update
				err = database.SetFeatureStatus(ctx, tx, "network", networkS)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "dns", dnsS2)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "gateway", gatewayS)
				g.Expect(err).To(BeNil())

				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(len(ss)).To(Equal(3))

				// network stayed the same
				g.Expect(ss["network"].Enabled).To(Equal(networkS.Enabled))
				g.Expect(ss["network"].Message).To(Equal(networkS.Message))
				g.Expect(ss["network"].Version).To(Equal(networkS.Version))
				g.Expect(ss["network"].UpdatedAt).To(Equal(networkS.UpdatedAt))

				// dns is updated
				g.Expect(ss["dns"].Enabled).To(Equal(dnsS2.Enabled))
				g.Expect(ss["dns"].Message).To(Equal(dnsS2.Message))
				g.Expect(ss["dns"].Version).To(Equal(dnsS2.Version))
				g.Expect(ss["dns"].UpdatedAt).To(Equal(dnsS2.UpdatedAt))

				// gateway is added
				g.Expect(ss["gateway"].Enabled).To(Equal(gatewayS.Enabled))
				g.Expect(ss["gateway"].Message).To(Equal(gatewayS.Message))
				g.Expect(ss["gateway"].Version).To(Equal(gatewayS.Version))
				g.Expect(ss["gateway"].UpdatedAt).To(Equal(gatewayS.UpdatedAt))
			})

			return nil
		})
	})
}
