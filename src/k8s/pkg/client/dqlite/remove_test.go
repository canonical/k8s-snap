package dqlite_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/client/dqlite"
	. "github.com/onsi/gomega"
)

func TestRemoveNodeByAddress(t *testing.T) {
	t.Run("Spare", func(t *testing.T) {
		withDqliteCluster(t, 2, func(ctx context.Context, dirs []string) {
			g := NewWithT(t)
			client, err := dqlite.NewClient(ctx, dqlite.ClientOpts{
				ClusterYAML: filepath.Join(dirs[0], "cluster.yaml"),
			})
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(client).NotTo(BeNil())

			members, err := client.ListMembers(ctx)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(members).To(HaveLen(2))

			memberToRemove := members[0].Address
			if members[0].Role == dqlite.Voter {
				memberToRemove = members[1].Address
			}
			g.Expect(client.RemoveNodeByAddress(ctx, memberToRemove)).To(Succeed())

			members, err = client.ListMembers(ctx)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(members).To(HaveLen(1))
		})
	})

	t.Run("LastVoter", func(t *testing.T) {
		withDqliteCluster(t, 2, func(ctx context.Context, dirs []string) {
			g := NewWithT(t)
			client, err := dqlite.NewClient(ctx, dqlite.ClientOpts{
				ClusterYAML: filepath.Join(dirs[0], "cluster.yaml"),
			})
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(client).NotTo(BeNil())

			members, err := client.ListMembers(ctx)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(members).To(HaveLen(2))

			memberToRemove := members[0]
			remainingNode := members[1]
			if members[0].Role != dqlite.Voter {
				memberToRemove = members[1]
				remainingNode = members[0]
			}
			g.Expect(memberToRemove.Role).To(Equal(dqlite.Voter))
			g.Expect(remainingNode.Role).To(Equal(dqlite.Spare))

			// Removing the last voter should succeed and leadership should be transferred.
			g.Expect(client.RemoveNodeByAddress(ctx, memberToRemove.Address)).To(Succeed())

			members, err = client.ListMembers(ctx)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(members).To(HaveLen(1))
			g.Expect(members[0].Role).To(Equal(dqlite.Voter))
			g.Expect(members[0].Address).ToNot(Equal(memberToRemove.Address))

			g.Expect(client.RemoveNodeByAddress(ctx, remainingNode.Address)).ToNot(Succeed())
		})
	})
}
