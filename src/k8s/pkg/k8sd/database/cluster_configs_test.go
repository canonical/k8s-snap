package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	. "github.com/onsi/gomega"
)

func TestDatabase(t *testing.T) {

	// TODO: We cannot split the tests because microcluster internally uses
	// global state that causes microcluster to fail if `WithDB` is called multiple times.
	t.Run("DatabaseTests", func(t *testing.T) {
		g := NewWithT(t)

		WithDB(t, func(ctx context.Context, d DB) {
			// Scenario 1: config can be written and retrieved from the database.
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

			// Scenario 2: Config keys can be  overwritten.
			expectedClusterConfig = database.ClusterConfig{
				K8sCertificateAuthority:    "some_cert",
				K8sCertificateAuthorityKey: "some_overwritten_key",
			}

			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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

			// Test kubernetes auth token valid token
			err = d.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				token, err := database.GetOrCreateToken(ctx, tx, "user1", []string{"group1", "group2"})
				if !g.Expect(err).To(BeNil()) {
					return err
				}
				g.Expect(token).To(Not(BeEmpty()))

				username, groups, err := database.CheckToken(ctx, tx, token)
				if !g.Expect(err).To(BeNil()) {
					return err
				}
				g.Expect(username).To(Equal("user1"))
				g.Expect(groups).To(ConsistOf("group1", "group2"))
				return nil
			})
			g.Expect(err).To(BeNil())
		})
	})
}
