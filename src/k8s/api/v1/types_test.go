package v1_test

import (
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestHaClusterFormed(t *testing.T) {
	g := NewWithT(t)

	testCases := []struct {
		name           string
		members        []apiv1.NodeStatus
		expectedResult bool
	}{
		{
			name: "Less than 3 voters",
			members: []apiv1.NodeStatus{
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleStandBy},
			},
			expectedResult: false,
		},
		{
			name: "Exactly 3 voters",
			members: []apiv1.NodeStatus{
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleVoter},
			},
			expectedResult: true,
		},
		{
			name: "More than 3 voters",
			members: []apiv1.NodeStatus{
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleVoter},
				{DatastoreRole: apiv1.DatastoreRoleStandBy},
			},
			expectedResult: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g.Expect(apiv1.ClusterStatus{Members: tc.members}.HaClusterFormed()).To(Equal(tc.expectedResult))
		})
	}
}

func TestString(t *testing.T) {
	testCases := []struct {
		name           string
		clusterStatus  apiv1.ClusterStatus
		expectedOutput string
	}{
		{
			name: "Cluster ready, HA formed, nodes exist",
			clusterStatus: apiv1.ClusterStatus{
				Ready: true,
				Members: []apiv1.NodeStatus{
					{Name: "node1", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.1"},
					{Name: "node2", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.2"},
					{Name: "node3", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.3"},
				},
				Config: apiv1.UserFacingClusterConfig{
					Network: apiv1.NetworkConfig{Enabled: utils.Pointer(true)},
					DNS:     apiv1.DNSConfig{Enabled: utils.Pointer(true)},
				},
				Datastore: apiv1.Datastore{Type: "k8s-dqlite"},
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
			clusterStatus: apiv1.ClusterStatus{
				Ready: true,
				Members: []apiv1.NodeStatus{
					{Name: "node1", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.1"},
				},
				Config: apiv1.UserFacingClusterConfig{
					Network: apiv1.NetworkConfig{Enabled: utils.Pointer(true)},
					DNS:     apiv1.DNSConfig{Enabled: utils.Pointer(true)},
				},
				Datastore: apiv1.Datastore{Type: "external", Servers: []string{"etcd-url1", "etcd-url2"}},
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
			clusterStatus: apiv1.ClusterStatus{
				Ready:     false,
				Members:   []apiv1.NodeStatus{},
				Config:    apiv1.UserFacingClusterConfig{},
				Datastore: apiv1.Datastore{},
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
