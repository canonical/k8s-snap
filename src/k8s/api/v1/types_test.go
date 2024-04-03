package v1

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

// This is expected to break if the default changes to make sure this is done intentionally.
func TestSetDefaults(t *testing.T) {
	g := NewWithT(t)

	b := &BootstrapConfig{}
	b.SetDefaults()

	expected := &BootstrapConfig{
		Components:    []string{"dns", "metrics-server", "network", "gateway"},
		ClusterCIDR:   "10.1.0.0/16",
		ServiceCIDR:   "10.152.183.0/24",
		EnableRBAC:    vals.Pointer(true),
		K8sDqlitePort: 9000,
		Datastore:     "k8s-dqlite",
	}

	g.Expect(b).To(Equal(expected))
}

func TestBootstrapConfigFromMap(t *testing.T) {
	g := NewWithT(t)
	// Create a new BootstrapConfig with default values
	bc := &BootstrapConfig{
		ClusterCIDR:   "10.1.0.0/16",
		Components:    []string{"dns", "network", "storage"},
		EnableRBAC:    vals.Pointer(true),
		K8sDqlitePort: 9000,
	}

	// Convert the BootstrapConfig to a map
	m, err := bc.ToMap()
	g.Expect(err).To(BeNil())

	// Unmarshal the YAML string from the map into a new BootstrapConfig instance
	bcyaml, err := BootstrapConfigFromMap(m)

	// Check for errors
	g.Expect(err).To(BeNil())
	// Compare the unmarshaled BootstrapConfig with the original one
	g.Expect(bcyaml).To(Equal(bc)) // Note the *bc here to compare values, not pointers

}

func TestHaClusterFormed(t *testing.T) {
	g := NewGomegaWithT(t)

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
	g := NewGomegaWithT(t)

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
					Network: &NetworkConfig{Enabled: vals.Pointer(true)},
					DNS:     &DNSConfig{Enabled: vals.Pointer(true)},
				},
				Datastore: Datastore{Type: "k8s-dqlite", ExternalURL: ""},
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
  cluster-domain: ""
  service-ip: ""
  upstream-nameservers: []
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
					Network: &NetworkConfig{Enabled: vals.Pointer(true)},
					DNS:     &DNSConfig{Enabled: vals.Pointer(true)},
				},
				Datastore: Datastore{Type: "external", ExternalURL: "I-am-a-postgres-url"},
			},
			expectedOutput: `status: ready
high-availability: no
datastore:
  type: external
  url: I-am-a-postgres-url

network:
  enabled: true
dns:
  enabled: true
  cluster-domain: ""
  service-ip: ""
  upstream-nameservers: []
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
			g.Expect(tc.clusterStatus.String()).To(Equal(tc.expectedOutput))
		})
	}
}
