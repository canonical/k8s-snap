package apiv1_test

import (
	"fmt"
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
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
					{Name: "node1", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.1", ClusterRole: apiv1.ClusterRoleControlPlane},
					{Name: "node2", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.2", ClusterRole: apiv1.ClusterRoleControlPlane},
					{Name: "node3", DatastoreRole: apiv1.DatastoreRoleStandBy, Address: "192.168.0.3", ClusterRole: apiv1.ClusterRoleControlPlane},
				},
				Datastore:    apiv1.Datastore{Type: "k8s-dqlite"},
				Network:      apiv1.FeatureStatus{Message: "enabled"},
				DNS:          apiv1.FeatureStatus{Message: "enabled at 192.168.0.10"},
				Ingress:      apiv1.FeatureStatus{Message: "enabled"},
				LoadBalancer: apiv1.FeatureStatus{Message: "enabled, L2 mode"},
				LocalStorage: apiv1.FeatureStatus{Message: "enabled at /var/snap/k8s/common/rawfile-storage"},
				Gateway:      apiv1.FeatureStatus{Message: "enabled"},
			},
			expectedOutput: `cluster status:           ready
control plane nodes:      192.168.0.1 (voter), 192.168.0.2 (voter), 192.168.0.3 (stand-by)
high availability:        no
datastore:                k8s-dqlite
network:                  enabled
dns:                      enabled at 192.168.0.10
ingress:                  enabled
load-balancer:            enabled, L2 mode
local-storage:            enabled at /var/snap/k8s/common/rawfile-storage
gateway                   enabled
`,
		},
		{
			name: "External Datastore",
			clusterStatus: apiv1.ClusterStatus{
				Ready: true,
				Members: []apiv1.NodeStatus{
					{Name: "node1", DatastoreRole: apiv1.DatastoreRoleVoter, Address: "192.168.0.1", ClusterRole: apiv1.ClusterRoleControlPlane},
				},
				Datastore: apiv1.Datastore{Type: "external", Servers: []string{"etcd-url1", "etcd-url2"}},
				Network:   apiv1.FeatureStatus{Message: "enabled"},
				DNS:       apiv1.FeatureStatus{Message: "enabled at 192.168.0.10"},
			},
			expectedOutput: `cluster status:           ready
control plane nodes:      192.168.0.1 (voter)
high availability:        no
datastore:                external
network:                  enabled
dns:                      enabled at 192.168.0.10
ingress:                  disabled
load-balancer:            disabled
local-storage:            disabled
gateway                   disabled
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
			expectedOutput: `cluster status:           not ready
control plane nodes:      none
high availability:        no
datastore:                disabled
network:                  disabled
dns:                      disabled
ingress:                  disabled
load-balancer:            disabled
local-storage:            disabled
gateway                   disabled
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			fmt.Println(tc.clusterStatus.String())
			g.Expect(tc.clusterStatus.String()).To(Equal(tc.expectedOutput))
		})
	}
}
