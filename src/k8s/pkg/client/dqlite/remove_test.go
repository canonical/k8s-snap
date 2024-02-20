package dqlite_test

import (
	"context"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/client/dqlite"
	. "github.com/onsi/gomega"
)

func TestRemoveNodeByAddress(t *testing.T) {
	t.Run("Spare", func(t *testing.T) {
		withDqliteCluster(t, 2, func(ctx context.Context, dirs []string) {
			g := NewWithT(t)
			client, err := dqlite.NewClient(ctx, dqlite.ClientOpts{
				ClusterYAML: path.Join(dirs[0], "cluster.yaml"),
			})
			g.Expect(err).To(BeNil())
			g.Expect(client).NotTo(BeNil())

			members, err := client.ListMembers(ctx)
			g.Expect(err).To(BeNil())
			g.Expect(members).To(HaveLen(2))

			memberToRemove := members[0].Address
			if members[0].Role == dqlite.Voter {
				memberToRemove = members[1].Address
			}
			g.Expect(client.RemoveNodeByAddress(ctx, memberToRemove)).To(BeNil())

			members, err = client.ListMembers(ctx)
			g.Expect(err).To(BeNil())
			g.Expect(members).To(HaveLen(1))
		})
	})

	t.Run("Voter", func(t *testing.T) {
		withDqliteCluster(t, 2, func(ctx context.Context, dirs []string) {
			g := NewWithT(t)
			client, err := dqlite.NewClient(ctx, dqlite.ClientOpts{
				ClusterYAML: path.Join(dirs[0], "cluster.yaml"),
			})
			g.Expect(err).To(BeNil())
			g.Expect(client).NotTo(BeNil())

			members, err := client.ListMembers(ctx)
			g.Expect(err).To(BeNil())
			g.Expect(members).To(HaveLen(2))

			memberToRemove := members[0].Address
			if members[0].Role != dqlite.Voter {
				memberToRemove = members[1].Address
			}
			g.Expect(client.RemoveNodeByAddress(ctx, memberToRemove)).To(BeNil())

			members, err = client.ListMembers(ctx)
			g.Expect(err).To(BeNil())
			g.Expect(members).To(HaveLen(1))
		})
	})
}
