package v1

import (
	"github.com/canonical/k8s/pkg/utils"
	"testing"

	. "github.com/onsi/gomega"
)

func TestHaClusterFormed(t *testing.T) {
	g := NewWithT(t)

	testCases := []struct {
		name           string
		members        []NodeStatus
		expectedResult bool
	}{
		{
			name: "Less than 3 voters",
			members: []NodeStatus{
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleStandBy},
			},
			expectedResult: false,
		},
		{
			name: "Exactly 3 voters",
			members: []NodeStatus{
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleVoter},
			},
			expectedResult: true,
		},
		{
			name: "More than 3 voters",
			members: []NodeStatus{
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleVoter},
				{DatastoreRole: DatastoreRoleStandBy},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g.Expect(ClusterStatus{Members: tc.members}.haClusterFormed()).To(Equal(tc.expectedResult))
		})
	}
}

func TestString(t *testing.T) {
	testCases := []struct {
		name           string
		clusterStatus  ClusterStatus
		expectedOutput string
	}{
		{
			name: "Cluster ready, HA formed, nodes exist",
			clusterStatus: ClusterStatus{
				Ready: true,
				Members: []NodeStatus{
					{Name: "node1", DatastoreRole: DatastoreRoleVoter, Address: "192.168.0.1"},
					{Name: "node2", DatastoreRole: DatastoreRoleVoter, Address: "192.168.0.2"},
					{Name: "node3", DatastoreRole: DatastoreRoleVoter, Address: "192.168.0.3"},
				},
				Config: UserFacingClusterConfig{
					Network: NetworkConfig{Enabled: utils.Pointer(true)},
					DNS:     DNSConfig{Enabled: utils.Pointer(true)},
				},
				Datastore: Datastore{Type: "k8s-dqlite"},
			},
			expectedOutput: `status: ready
high-availability: yes
datastore:
  type: k8s-dqlite
  voter-nodes:
    - 192.168.0.1
    - 192.168.0.2
    - 192.168.0.3
  standby-nodes: none
  spare-nodes: none
network:
  enabled: true
dns:
  enabled: true
`,
		},
		{
			name: "External Datastore",
			clusterStatus: ClusterStatus{
				Ready: true,
				Members: []NodeStatus{
					{Name: "node1", DatastoreRole: DatastoreRoleVoter, Address: "192.168.0.1"},
				},
				Config: UserFacingClusterConfig{
					Network: NetworkConfig{Enabled: utils.Pointer(true)},
					DNS:     DNSConfig{Enabled: utils.Pointer(true)},
				},
				Datastore: Datastore{Type: "external", Servers: []string{"etcd-url1", "etcd-url2"}},
			},
			expectedOutput: `status: ready
high-availability: no
datastore:
  type: external
  servers:
    - etcd-url1
    - etcd-url2
network:
  enabled: true
dns:
  enabled: true
`,
		},
		{
			name: "Cluster not ready, HA not formed, no nodes",
			clusterStatus: ClusterStatus{
				Ready:     false,
				Members:   []NodeStatus{},
				Config:    UserFacingClusterConfig{},
				Datastore: Datastore{},
			},
			expectedOutput: `status: not ready
high-availability: no
datastore:
  voter-nodes: none
  standby-nodes: none
  spare-nodes: none
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tc.clusterStatus.String()).To(Equal(tc.expectedOutput))
		})
	}
}
