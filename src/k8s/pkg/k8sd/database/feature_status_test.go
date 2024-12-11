package database_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	microcluster_testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

func TestFeatureStatus(t *testing.T) {
	microcluster_testenv.WithState(t, func(ctx context.Context, s state.State) {
		_ = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(ss).To(BeEmpty())
			})

			t.Run("SettingNewStatus", func(t *testing.T) {
				g := NewWithT(t)

				err := database.SetFeatureStatus(ctx, tx, features.Network, networkStatus)
				g.Expect(err).To(Not(HaveOccurred()))
				err = database.SetFeatureStatus(ctx, tx, features.DNS, dnsStatus)
				g.Expect(err).To(Not(HaveOccurred()))

				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(ss).To(HaveLen(2))

				g.Expect(ss[features.Network].Enabled).To(Equal(networkStatus.Enabled))
				g.Expect(ss[features.Network].Message).To(Equal(networkStatus.Message))
				g.Expect(ss[features.Network].Version).To(Equal(networkStatus.Version))
				g.Expect(ss[features.Network].UpdatedAt).To(Equal(networkStatus.UpdatedAt))

				g.Expect(ss[features.DNS].Enabled).To(Equal(dnsStatus.Enabled))
				g.Expect(ss[features.DNS].Message).To(Equal(dnsStatus.Message))
				g.Expect(ss[features.DNS].Version).To(Equal(dnsStatus.Version))
				g.Expect(ss[features.DNS].UpdatedAt).To(Equal(dnsStatus.UpdatedAt))
			})
			t.Run("UpdatingStatus", func(t *testing.T) {
				g := NewWithT(t)

				err := database.SetFeatureStatus(ctx, tx, features.Network, networkStatus)
				g.Expect(err).To(Not(HaveOccurred()))
				err = database.SetFeatureStatus(ctx, tx, features.DNS, dnsStatus)
				g.Expect(err).To(Not(HaveOccurred()))

				// set and update
				err = database.SetFeatureStatus(ctx, tx, features.Network, networkStatus)
				g.Expect(err).To(Not(HaveOccurred()))
				err = database.SetFeatureStatus(ctx, tx, features.DNS, dnsStatus2)
				g.Expect(err).To(Not(HaveOccurred()))
				err = database.SetFeatureStatus(ctx, tx, features.Gateway, gatewayStatus)
				g.Expect(err).To(Not(HaveOccurred()))

				ss, err := database.GetFeatureStatuses(ctx, tx)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(ss).To(HaveLen(3))

				// network stayed the same
				g.Expect(ss[features.Network].Enabled).To(Equal(networkStatus.Enabled))
				g.Expect(ss[features.Network].Message).To(Equal(networkStatus.Message))
				g.Expect(ss[features.Network].Version).To(Equal(networkStatus.Version))
				g.Expect(ss[features.Network].UpdatedAt).To(Equal(networkStatus.UpdatedAt))

				// dns is updated
				g.Expect(ss[features.DNS].Enabled).To(Equal(dnsStatus2.Enabled))
				g.Expect(ss[features.DNS].Message).To(Equal(dnsStatus2.Message))
				g.Expect(ss[features.DNS].Version).To(Equal(dnsStatus2.Version))
				g.Expect(ss[features.DNS].UpdatedAt).To(Equal(dnsStatus2.UpdatedAt))

				// gateway is added
				g.Expect(ss[features.Gateway].Enabled).To(Equal(gatewayStatus.Enabled))
				g.Expect(ss[features.Gateway].Message).To(Equal(gatewayStatus.Message))
				g.Expect(ss[features.Gateway].Version).To(Equal(gatewayStatus.Version))
				g.Expect(ss[features.Gateway].UpdatedAt).To(Equal(gatewayStatus.UpdatedAt))
			})

			return nil
		})
	})
}
