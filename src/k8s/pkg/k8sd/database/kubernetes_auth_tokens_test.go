package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	. "github.com/onsi/gomega"
)

func TestKubernetesAuthTokens(t *testing.T) {
	WithDB(t, func(ctx context.Context, db DB) {
		var token1, token2 string

		t.Run("GetOrCreateToken", func(t *testing.T) {
			g := NewWithT(t)
			err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				var err error

				token1, err = database.GetOrCreateToken(ctx, tx, "user1", []string{"group1", "group2"})
				g.Expect(err).To(BeNil())
				g.Expect(token1).To(Not(BeEmpty()))

				token2, err = database.GetOrCreateToken(ctx, tx, "user2", []string{"group1", "group2"})
				g.Expect(err).To(BeNil())
				g.Expect(token2).To(Not(BeEmpty()))

				g.Expect(token1).To(Not(Equal(token2)))
				return nil
			})
			g.Expect(err).To(BeNil())

			t.Run("Existing", func(t *testing.T) {
				g := NewWithT(t)
				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					token, err := database.GetOrCreateToken(ctx, tx, "user1", []string{"group1", "group2"})
					g.Expect(err).To(BeNil())
					g.Expect(token).To(Equal(token1))
					return nil
				})
				g.Expect(err).To(BeNil())
			})
		})

		t.Run("CheckToken", func(t *testing.T) {
			t.Run("user1", func(t *testing.T) {
				g := NewWithT(t)
				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					username, groups, err := database.CheckToken(ctx, tx, token1)
					g.Expect(err).To(BeNil())
					g.Expect(username).To(Equal("user1"))
					g.Expect(groups).To(ConsistOf("group1", "group2"))
					return nil
				})
				g.Expect(err).To(BeNil())
			})
			t.Run("user2", func(t *testing.T) {
				g := NewWithT(t)
				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					username, groups, err := database.CheckToken(ctx, tx, token2)
					g.Expect(err).To(BeNil())
					g.Expect(username).To(Equal("user2"))
					g.Expect(groups).To(ConsistOf("group1", "group2"))
					return nil
				})
				g.Expect(err).To(BeNil())
			})
		})

		t.Run("DeleteToken", func(t *testing.T) {
			g := NewWithT(t)
			err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.DeleteToken(ctx, tx, token2)
				g.Expect(err).To(BeNil())

				username, groups, err := database.CheckToken(ctx, tx, token2)
				g.Expect(err).ToNot(BeNil())
				g.Expect(username).To(BeEmpty())
				g.Expect(groups).To(BeEmpty())
				return nil
			})
			g.Expect(err).To(BeNil())
		})
	})
}
