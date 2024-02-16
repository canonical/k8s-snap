package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestClusterConfig(t *testing.T) {
	WithDB(t, func(ctx context.Context, d DB) {
		t.Run("Set", func(t *testing.T) {
			g := NewWithT(t)
			expectedClusterConfig := types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: "CA CERT DATA",
					CAKey:  "CA KEY DATA",
				},
			}

			// Write some config to the database
			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.SetClusterConfig(context.Background(), tx, expectedClusterConfig)
				g.Expect(err).To(BeNil())
				return nil
			})
			g.Expect(err).To(BeNil())

			// Retrieve it and map it to the struct
			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(BeNil())
		})

		t.Run("CannotUpdateCA", func(t *testing.T) {
			// TODO(neoaggelos): extend this test for all fields that cannot be updated
			g := NewWithT(t)
			expectedClusterConfig := types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: "CA CERT DATA",
					CAKey:  "CA KEY DATA",
				},
			}

			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.SetClusterConfig(context.Background(), tx, types.ClusterConfig{
					Certificates: types.Certificates{
						CACert: "CA CERT NEW DATA",
					},
				})
				g.Expect(err).To(HaveOccurred())
				return err
			})
			g.Expect(err).To(HaveOccurred())

			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(BeNil())
		})

		t.Run("Update", func(t *testing.T) {
			g := NewWithT(t)
			expectedClusterConfig := types.ClusterConfig{
				Certificates: types.Certificates{
					CACert:        "CA CERT DATA",
					CAKey:         "CA KEY DATA",
					K8sDqliteCert: "CERT DATA",
					K8sDqliteKey:  "KEY DATA",
				},
				Kubelet: types.Kubelet{
					ClusterDNS: "10.152.183.10",
				},
				APIServer: types.APIServer{
					ServiceAccountKey: "SA KEY DATA",
				},
			}

			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.SetClusterConfig(context.Background(), tx, types.ClusterConfig{
					Kubelet: types.Kubelet{
						ClusterDNS: "10.152.183.10",
					},
					Certificates: types.Certificates{
						K8sDqliteCert: "CERT DATA",
						K8sDqliteKey:  "KEY DATA",
					},
					APIServer: types.APIServer{
						ServiceAccountKey: "SA KEY DATA",
					},
				})
				g.Expect(err).To(BeNil())
				return nil
			})
			g.Expect(err).To(BeNil())

			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				clusterConfig, err := database.GetClusterConfig(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(clusterConfig).To(Equal(expectedClusterConfig))
				return nil
			})
			g.Expect(err).To(BeNil())
		})
	})
}
