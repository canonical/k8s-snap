package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func TestClusterConfig(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
		t.Run("Set", func(t *testing.T) {
			g := NewWithT(t)
			expectedClusterConfig := types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: utils.Pointer("CA CERT DATA"),
					CAKey:  utils.Pointer("CA KEY DATA"),
				},
			}
			expectedClusterConfig.SetDefaults()

			// Write some config to the database
			err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				_, err := database.SetClusterConfig(context.Background(), tx, expectedClusterConfig)
				g.Expect(err).To(Not(HaveOccurred()))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))

			// Retrieve it and map it to the struct
			err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))
		})

		t.Run("CannotUpdateCA", func(t *testing.T) {
			// TODO(neoaggelos): extend this test for all fields that cannot be updated
			g := NewWithT(t)
			expectedClusterConfig := types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: utils.Pointer("CA CERT DATA"),
					CAKey:  utils.Pointer("CA KEY DATA"),
				},
			}
			expectedClusterConfig.SetDefaults()

			err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				_, err := database.SetClusterConfig(context.Background(), tx, types.ClusterConfig{
					Certificates: types.Certificates{
						CACert: utils.Pointer("CA CERT NEW DATA"),
					},
				})
				g.Expect(err).To(HaveOccurred())
				return err
			})
			g.Expect(err).To(HaveOccurred())

			err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))
		})

		t.Run("Update", func(t *testing.T) {
			g := NewWithT(t)
			expectedClusterConfig := types.ClusterConfig{
				Certificates: types.Certificates{
					CACert:            utils.Pointer("CA CERT DATA"),
					CAKey:             utils.Pointer("CA KEY DATA"),
					ServiceAccountKey: utils.Pointer("SA KEY DATA"),
				},
				Datastore: types.Datastore{
					K8sDqliteCert: utils.Pointer("CERT DATA"),
					K8sDqliteKey:  utils.Pointer("KEY DATA"),
				},
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("10.152.183.10"),
				},
			}
			expectedClusterConfig.SetDefaults()

			err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				returnedConfig, err := database.SetClusterConfig(context.Background(), tx, types.ClusterConfig{
					Kubelet: types.Kubelet{
						ClusterDNS: utils.Pointer("10.152.183.10"),
					},
					Datastore: types.Datastore{
						K8sDqliteCert: utils.Pointer("CERT DATA"),
						K8sDqliteKey:  utils.Pointer("KEY DATA"),
					},
					Certificates: types.Certificates{
						ServiceAccountKey: utils.Pointer("SA KEY DATA"),
					},
				})
				g.Expect(returnedConfig).To(Equal(expectedClusterConfig))
				g.Expect(err).To(Not(HaveOccurred()))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))

			err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))
		})

		t.Run("SetBootstrapConfig", func(t *testing.T) {
			g := NewWithT(t)
			expBootstrapConfig := types.ClusterConfig{}
			expBootstrapConfig.SetDefaults()

			err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				return database.SetClusterBootstrapConfig(context.Background(), tx, expBootstrapConfig)
			})
			g.Expect(err).To(Not(HaveOccurred()))

			err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				bootstrapConfig, err := database.GetClusterBootstrapConfig(ctx, tx)
				g.Expect(bootstrapConfig).To(Equal(expBootstrapConfig))
				return err
			})
			g.Expect(err).To(Not(HaveOccurred()))

			newConfig := expBootstrapConfig
			// Toggle the network enabled field as an example of a change
			if *newConfig.Network.Enabled {
				newConfig.Network.Enabled = ptr.To(false)
			} else {
				newConfig.Network.Enabled = ptr.To(true)
			}

			err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				return database.SetClusterBootstrapConfig(context.Background(), tx, newConfig)
			})
			g.Expect(err).To(Not(HaveOccurred()))

			err = s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				bootstrapConfig, err := database.GetClusterBootstrapConfig(ctx, tx)
				g.Expect(bootstrapConfig).To(Equal(expBootstrapConfig))
				return err
			})
			g.Expect(err).To(Not(HaveOccurred()))
		})
	})
}
