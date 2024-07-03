package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	. "github.com/onsi/gomega"
)

func TestClusterAPIAuthTokens(t *testing.T) {
	WithDB(t, func(ctx context.Context, db DB) {
		var token string = "test-token"

		t.Run("SetAuthToken", func(t *testing.T) {
			g := NewWithT(t)
			err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.SetClusterAPIToken(ctx, tx, token)
				g.Expect(err).To(BeNil())
				return nil
			})
			g.Expect(err).To(BeNil())
		})

		t.Run("CheckAuthToken", func(t *testing.T) {
			t.Run("ValidToken", func(t *testing.T) {
				g := NewWithT(t)
				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					valid, err := database.ValidateClusterAPIToken(ctx, tx, token)
					g.Expect(err).To(BeNil())
					g.Expect(valid).To(BeTrue())
					return nil
				})
				g.Expect(err).To(BeNil())
			})

			t.Run("InvalidToken", func(t *testing.T) {
				g := NewWithT(t)
				err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					valid, err := database.ValidateClusterAPIToken(ctx, tx, "invalid-token")
					g.Expect(err).To(BeNil())
					g.Expect(valid).To(BeFalse())
					return nil
				})
				g.Expect(err).To(BeNil())
			})
		})
	})
}
