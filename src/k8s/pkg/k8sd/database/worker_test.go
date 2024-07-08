package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/database"
	. "github.com/onsi/gomega"
)

func TestWorkerNodeToken(t *testing.T) {
	WithDB(t, func(ctx context.Context, db DB) {
		_ = db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			t.Run("Default", func(t *testing.T) {
				g := NewWithT(t)
				exists, err := database.CheckWorkerNodeToken(ctx, tx, "somenode", "sometoken")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeFalse())

				token, err := database.GetOrCreateWorkerNodeToken(ctx, tx, "somenode")
				g.Expect(err).To(BeNil())
				g.Expect(token).To(HaveLen(48))

				othertoken, err := database.GetOrCreateWorkerNodeToken(ctx, tx, "someothernode")
				g.Expect(err).To(BeNil())
				g.Expect(othertoken).To(HaveLen(48))
				g.Expect(othertoken).NotTo(Equal(token))

				valid, err := database.CheckWorkerNodeToken(ctx, tx, "somenode", token)
				g.Expect(err).To(BeNil())
				g.Expect(valid).To(BeTrue())

				valid, err = database.CheckWorkerNodeToken(ctx, tx, "someothernode", token)
				g.Expect(err).To(BeNil())
				g.Expect(valid).To(BeFalse())

				valid, err = database.CheckWorkerNodeToken(ctx, tx, "someothernode", othertoken)
				g.Expect(err).To(BeNil())
				g.Expect(valid).To(BeTrue())

				err = database.DeleteWorkerNodeToken(ctx, tx, token)
				g.Expect(err).To(BeNil())

				valid, err = database.CheckWorkerNodeToken(ctx, tx, "somenode", token)
				g.Expect(err).To(BeNil())
				g.Expect(valid).To(BeFalse())

				newToken, err := database.GetOrCreateWorkerNodeToken(ctx, tx, "somenode")
				g.Expect(err).To(BeNil())
				g.Expect(newToken).To(HaveLen(48))
				g.Expect(newToken).ToNot(Equal(token))
			})

			t.Run("AnyNodeName", func(t *testing.T) {
				g := NewWithT(t)
				token, err := database.GetOrCreateWorkerNodeToken(ctx, tx, "")
				g.Expect(err).To(BeNil())
				g.Expect(token).To(HaveLen(48))

				for _, name := range []string{"", "test", "other"} {
					t.Run(name, func(t *testing.T) {
						g := NewWithT(t)

						valid, err := database.CheckWorkerNodeToken(ctx, tx, name, token)
						g.Expect(err).To(BeNil())
						g.Expect(valid).To(BeTrue())
					})
				}
			})
			return nil
		})
	})
}
