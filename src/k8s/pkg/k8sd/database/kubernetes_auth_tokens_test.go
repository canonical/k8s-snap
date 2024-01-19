package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	. "github.com/onsi/gomega"
)

func TestKubernetesAuthTokens(t *testing.T) {
	t.Run("ValidToken", func(t *testing.T) {
		g := NewWithT(t)
		WithDB(t, func(ctx context.Context, db DB) {
			err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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
