package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	testenv "github.com/canonical/k8s/pkg/utils/microcluster"
	"github.com/canonical/microcluster/v2/state"
	. "github.com/onsi/gomega"
)

func TestClusterAPIAuthTokens(t *testing.T) {
	testenv.WithState(t, func(ctx context.Context, s state.State) {
		var token string = "test-token"

		t.Run("SetAuthToken", func(t *testing.T) {
			g := NewWithT(t)
			err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
				err := database.SetClusterAPIToken(ctx, tx, token)
				g.Expect(err).To(Not(HaveOccurred()))
				return nil
			})
			g.Expect(err).To(Not(HaveOccurred()))
		})

		t.Run("CheckAuthToken", func(t *testing.T) {
			t.Run("ValidToken", func(t *testing.T) {
				g := NewWithT(t)
				err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					valid, err := database.ValidateClusterAPIToken(ctx, tx, token)
					g.Expect(err).To(Not(HaveOccurred()))
					g.Expect(valid).To(BeTrue())
					return nil
				})
				g.Expect(err).To(Not(HaveOccurred()))
			})

			t.Run("InvalidToken", func(t *testing.T) {
				g := NewWithT(t)
				err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
					valid, err := database.ValidateClusterAPIToken(ctx, tx, "invalid-token")
					g.Expect(err).To(Not(HaveOccurred()))
					g.Expect(valid).To(BeFalse())
					return nil
				})
				g.Expect(err).To(Not(HaveOccurred()))
			})
		})
	})
}
