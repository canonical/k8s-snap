package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestClusterConfig(t *testing.T) {
	WithDB(t, func(ctx context.Context, d DB) {
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
			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				_, err := database.SetClusterConfig(context.Background(), tx, expectedClusterConfig)
				g.Expect(err).To(Not(HaveOccurred()))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))

			// Retrieve it and map it to the struct
			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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

			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				_, err := database.SetClusterConfig(context.Background(), tx, types.ClusterConfig{
					Certificates: types.Certificates{
						CACert: utils.Pointer("CA CERT NEW DATA"),
					},
				})
				g.Expect(err).To(HaveOccurred())
				return err
			})
			g.Expect(err).To(HaveOccurred())

			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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

			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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

			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))
		})
	})
}
