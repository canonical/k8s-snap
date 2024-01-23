package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	. "github.com/onsi/gomega"
)

func TestClusterConfig(t *testing.T) {
	WithDB(t, func(ctx context.Context, d DB) {
		t.Run("Set", func(t *testing.T) {
			g := NewWithT(t)
			expectedClusterConfig := database.ClusterConfig{
				K8sCertificateAuthority:    "some_cert",
				K8sCertificateAuthorityKey: "some_key",
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

		t.Run("Update", func(t *testing.T) {
			g := NewWithT(t)
			expectedClusterConfig := database.ClusterConfig{
				K8sCertificateAuthority:    "some_cert",
				K8sCertificateAuthorityKey: "some_overwritten_key",
			}

			err := d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.SetClusterConfig(context.Background(), tx, expectedClusterConfig)
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
