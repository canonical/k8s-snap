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
			networkStatus := types.FeatureStatus{
				Enabled:   true,
				Message:   "enabled",
				Version:   "1.2.3",
				UpdatedAt: t0,
			}
			dnsStatus := types.FeatureStatus{
				Enabled:   true,
				Message:   "enabled at 10.0.0.1",
				Version:   "4.5.6",
				UpdatedAt: t0,
			}
			dnsStatus2 := types.FeatureStatus{
				Enabled:   true,
				Message:   "enabled at 10.0.0.2",
				Version:   "4.5.7",
				UpdatedAt: t0,
			}
			gatewayStatus := types.FeatureStatus{
				Enabled:   true,
				Message:   "disabled",
				Version:   "10.20.30",
				UpdatedAt: t0,
			}

			t.Run("ReturnNothingInitially", func(t *testing.T) {
				g := NewWithT(t)
				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(ss).To(BeEmpty())

			})

			t.Run("SettingNewStatus", func(t *testing.T) {
				g := NewWithT(t)

				err := database.SetFeatureStatus(ctx, tx, "network", networkStatus)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "dns", dnsStatus)
				g.Expect(err).To(BeNil())

				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(ss).To(HaveLen(2))

				g.Expect(ss["network"].Enabled).To(Equal(networkStatus.Enabled))
				g.Expect(ss["network"].Message).To(Equal(networkStatus.Message))
				g.Expect(ss["network"].Version).To(Equal(networkStatus.Version))
				g.Expect(ss["network"].UpdatedAt).To(Equal(networkStatus.UpdatedAt))

				g.Expect(ss["dns"].Enabled).To(Equal(dnsStatus.Enabled))
				g.Expect(ss["dns"].Message).To(Equal(dnsStatus.Message))
				g.Expect(ss["dns"].Version).To(Equal(dnsStatus.Version))
				g.Expect(ss["dns"].UpdatedAt).To(Equal(dnsStatus.UpdatedAt))

			})
			t.Run("UpdatingStatus", func(t *testing.T) {
				g := NewWithT(t)

				err := database.SetFeatureStatus(ctx, tx, "network", networkStatus)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "dns", dnsStatus)
				g.Expect(err).To(BeNil())

				// set and update
				err = database.SetFeatureStatus(ctx, tx, "network", networkStatus)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "dns", dnsStatus2)
				g.Expect(err).To(BeNil())
				err = database.SetFeatureStatus(ctx, tx, "gateway", gatewayStatus)
				g.Expect(err).To(BeNil())

				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(ss).To(HaveLen(3))

				// network stayed the same
				g.Expect(ss["network"].Enabled).To(Equal(networkStatus.Enabled))
				g.Expect(ss["network"].Message).To(Equal(networkStatus.Message))
				g.Expect(ss["network"].Version).To(Equal(networkStatus.Version))
				g.Expect(ss["network"].UpdatedAt).To(Equal(networkStatus.UpdatedAt))

				// dns is updated
				g.Expect(ss["dns"].Enabled).To(Equal(dnsStatus2.Enabled))
				g.Expect(ss["dns"].Message).To(Equal(dnsStatus2.Message))
				g.Expect(ss["dns"].Version).To(Equal(dnsStatus2.Version))
				g.Expect(ss["dns"].UpdatedAt).To(Equal(dnsStatus2.UpdatedAt))

				// gateway is added
				g.Expect(ss["gateway"].Enabled).To(Equal(gatewayStatus.Enabled))
				g.Expect(ss["gateway"].Message).To(Equal(gatewayStatus.Message))
				g.Expect(ss["gateway"].Version).To(Equal(gatewayStatus.Version))
				g.Expect(ss["gateway"].UpdatedAt).To(Equal(gatewayStatus.UpdatedAt))
			})

			return nil
		})
	})
}
