package etcd

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// mockCluster implements clientv3.Cluster. Embedded interface satisfies
// methods we don't use; only MemberList is overridden.
type mockCluster struct {
	clientv3.Cluster
	members       []*pb.Member
	memberListErr error
}

func (m *mockCluster) MemberList(_ context.Context, _ ...clientv3.OpOption) (*clientv3.MemberListResponse, error) {
	if m.memberListErr != nil {
		return nil, m.memberListErr
	}

	return (*clientv3.MemberListResponse)(&pb.MemberListResponse{Members: m.members}), nil
}

// mockMaintenance implements clientv3.Maintenance. Only Status and MoveLeader
// are overridden; unused methods panic via the embedded nil interface.
type mockMaintenance struct {
	clientv3.Maintenance
	leaderID             uint64
	statusErr            error
	moveLeaderErr        error
	moveLeaderCalledWith uint64
}

func (m *mockMaintenance) Status(_ context.Context, _ string) (*clientv3.StatusResponse, error) {
	if m.statusErr != nil {
		return nil, m.statusErr
	}

	return (*clientv3.StatusResponse)(&pb.StatusResponse{Leader: m.leaderID}), nil
}

func (m *mockMaintenance) MoveLeader(_ context.Context, transfereeID uint64) (*clientv3.MoveLeaderResponse, error) {
	m.moveLeaderCalledWith = transfereeID
	return nil, m.moveLeaderErr
}

func newTestClient(cluster *mockCluster, maintenance *mockMaintenance) *Client {
	c := clientv3.NewCtxClient(context.Background())
	c.Cluster = cluster
	c.Maintenance = maintenance

	return &Client{Client: c}
}

func mockFactory(maintenance *mockMaintenance) func(string, []string) (*Client, error) {
	return func(_ string, _ []string) (*Client, error) {
		return newTestClient(nil, maintenance), nil
	}
}

func TestMoveLeaderIfNeeded(t *testing.T) {
	nodeA := &pb.Member{ID: 1, Name: "node-a", ClientURLs: []string{"https://10.0.0.1:2379"}}
	nodeB := &pb.Member{ID: 2, Name: "node-b", ClientURLs: []string{"https://10.0.0.2:2379"}}
	nodeBLearner := &pb.Member{ID: 2, Name: "node-b", ClientURLs: []string{"https://10.0.0.2:2379"}, IsLearner: true}
	nodeC := &pb.Member{ID: 3, Name: "node-c", ClientURLs: []string{"https://10.0.0.3:2379"}}

	t.Run("NodeNotInMemberList", func(t *testing.T) {
		g := NewWithT(t)
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeB}},
			&mockMaintenance{leaderID: nodeA.ID},
		)
		g.Expect(client.moveLeader(context.Background(), "ghost-node", nil)).To(Succeed())
	})

	t.Run("NodeIsNotLeader", func(t *testing.T) {
		g := NewWithT(t)
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeB}},
			&mockMaintenance{leaderID: nodeB.ID},
		)
		g.Expect(client.moveLeader(context.Background(), nodeA.Name, nil)).To(Succeed())
	})

	t.Run("NodeIsLeaderTransfersToVoter", func(t *testing.T) {
		g := NewWithT(t)
		leaderMaint := &mockMaintenance{}
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeB, nodeC}},
			&mockMaintenance{leaderID: nodeA.ID},
		)
		g.Expect(client.moveLeader(context.Background(), nodeA.Name, mockFactory(leaderMaint))).To(Succeed())
		g.Expect(leaderMaint.moveLeaderCalledWith).To(Equal(nodeB.ID))
	})

	t.Run("SkipsLearnerWhenChoosingTransferee", func(t *testing.T) {
		g := NewWithT(t)
		leaderMaint := &mockMaintenance{}
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeBLearner, nodeC}},
			&mockMaintenance{leaderID: nodeA.ID},
		)
		g.Expect(client.moveLeader(context.Background(), nodeA.Name, mockFactory(leaderMaint))).To(Succeed())
		g.Expect(leaderMaint.moveLeaderCalledWith).To(Equal(nodeC.ID))
	})

	t.Run("MemberListError", func(t *testing.T) {
		g := NewWithT(t)
		client := newTestClient(
			&mockCluster{memberListErr: errors.New("etcd unreachable")},
			&mockMaintenance{},
		)
		err := client.moveLeader(context.Background(), nodeA.Name, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to list etcd members"))
	})

	t.Run("AllStatusCallsFail", func(t *testing.T) {
		g := NewWithT(t)
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeB}},
			&mockMaintenance{statusErr: errors.New("timeout")},
		)
		err := client.moveLeader(context.Background(), nodeA.Name, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to determine etcd leader"))
	})

	t.Run("LeaderHasNoClientURLs", func(t *testing.T) {
		g := NewWithT(t)
		leaderNoURLs := &pb.Member{ID: 1, Name: "node-a"}
		client := newTestClient(
			&mockCluster{members: []*pb.Member{leaderNoURLs, nodeB}},
			&mockMaintenance{leaderID: leaderNoURLs.ID},
		)
		err := client.moveLeader(context.Background(), leaderNoURLs.Name, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to determine etcd leader"))
	})

	t.Run("NoVoterAvailableForTransfer", func(t *testing.T) {
		g := NewWithT(t)
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeBLearner}},
			&mockMaintenance{leaderID: nodeA.ID},
		)
		err := client.moveLeader(context.Background(), nodeA.Name, nil)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("no eligible etcd member"))
	})

	t.Run("MoveLeaderError", func(t *testing.T) {
		g := NewWithT(t)
		leaderMaint := &mockMaintenance{moveLeaderErr: errors.New("not leader")}
		client := newTestClient(
			&mockCluster{members: []*pb.Member{nodeA, nodeB}},
			&mockMaintenance{leaderID: nodeA.ID},
		)
		err := client.moveLeader(context.Background(), nodeA.Name, mockFactory(leaderMaint))
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to transfer etcd leadership"))
	})
}
