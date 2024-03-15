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
		g := NewWithT(t)
		err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
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

			err = database.DeleteWorkerNodeToken(ctx, tx, "somenode")
			g.Expect(err).To(BeNil())

			valid, err = database.CheckWorkerNodeToken(ctx, tx, "somenode", token)
			g.Expect(err).To(BeNil())
			g.Expect(valid).To(BeFalse())

			newToken, err := database.GetOrCreateWorkerNodeToken(ctx, tx, "somenode")
			g.Expect(err).To(BeNil())
			g.Expect(newToken).To(HaveLen(48))
			g.Expect(newToken).ToNot(Equal(token))
			return nil
		})
		g.Expect(err).To(BeNil())
	})
}

func TestWorkerNodes(t *testing.T) {
	WithDB(t, func(ctx context.Context, db DB) {
		g := NewWithT(t)
		err := db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			t.Run("Empty", func(t *testing.T) {
				g := NewWithT(t)

				nodes, err := database.ListWorkerNodes(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(nodes).To(BeEmpty())
			})

			t.Run("AddOne", func(t *testing.T) {
				g := NewWithT(t)

				err := database.AddWorkerNode(ctx, tx, "w1")
				g.Expect(err).To(BeNil())

				nodes, err := database.ListWorkerNodes(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(nodes).To(ConsistOf("w1"))

				exists, err := database.CheckWorkerExists(ctx, tx, "w1")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeTrue())
			})

			t.Run("AddTwo", func(t *testing.T) {
				g := NewWithT(t)

				err := database.AddWorkerNode(ctx, tx, "w2")
				g.Expect(err).To(BeNil())

				nodes, err := database.ListWorkerNodes(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(nodes).To(ConsistOf("w1", "w2"))

				exists, err := database.CheckWorkerExists(ctx, tx, "w1")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeTrue())

				exists, err = database.CheckWorkerExists(ctx, tx, "w2")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeTrue())
			})

			t.Run("AddDuplicateFails", func(t *testing.T) {
				g := NewWithT(t)

				err := database.AddWorkerNode(ctx, tx, "w1")
				g.Expect(err).To(HaveOccurred())

				nodes, err := database.ListWorkerNodes(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(nodes).To(ConsistOf("w1", "w2"))
			})

			t.Run("Delete", func(t *testing.T) {
				g := NewWithT(t)

				err := database.DeleteWorkerNode(ctx, tx, "w1")
				g.Expect(err).To(BeNil())

				nodes, err := database.ListWorkerNodes(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(nodes).To(ConsistOf("w2"))

				exists, err := database.CheckWorkerExists(ctx, tx, "w1")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeFalse())

				exists, err = database.CheckWorkerExists(ctx, tx, "w2")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeTrue())
			})

			t.Run("ReuseName", func(t *testing.T) {
				g := NewWithT(t)

				err := database.AddWorkerNode(ctx, tx, "w1")
				g.Expect(err).To(BeNil())

				nodes, err := database.ListWorkerNodes(ctx, tx)
				g.Expect(err).To(BeNil())
				g.Expect(nodes).To(ConsistOf("w1", "w2"))

				exists, err := database.CheckWorkerExists(ctx, tx, "w1")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeTrue())

				exists, err = database.CheckWorkerExists(ctx, tx, "w2")
				g.Expect(err).To(BeNil())
				g.Expect(exists).To(BeTrue())
			})
			return nil
		})
		g.Expect(err).To(BeNil())
	})
}
