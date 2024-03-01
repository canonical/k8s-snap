package v1

import (
	"fmt"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
)

type BootstrapConfig struct {
	// Components are the components that should be enabled on bootstrap.
	Components []string `yaml:"components"`
	// ClusterCIDR is the CIDR of the cluster.
	ClusterCIDR string `yaml:"cluster-cidr"`
	// EnableRBAC determines if RBAC will be enabled; *bool to know true/false/unset.
	EnableRBAC    *bool `yaml:"enable-rbac"`
	K8sDqlitePort int   `yaml:"k8s-dqlite-port"`
}

// SetDefaults sets the fields to default values.
func (b *BootstrapConfig) SetDefaults() {
	b.Components = []string{"dns", "network"}
	b.ClusterCIDR = "10.1.0.0/16"
	b.EnableRBAC = &[]bool{true}[0]
	b.K8sDqlitePort = 9000
}

// ToMap marshals the BootstrapConfig into yaml and map it to "bootstrapConfig".
func (b *BootstrapConfig) ToMap() (map[string]string, error) {
	config, err := yaml.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config map: %w", err)
	}

	return map[string]string{
		"bootstrapConfig": string(config),
	}, nil
}

// BootstrapConfigFromMap converts a string map to a BootstrapConfig struct.
func BootstrapConfigFromMap(m map[string]string) (*BootstrapConfig, error) {
	config := &BootstrapConfig{}
	err := yaml.Unmarshal([]byte(m["bootstrapConfig"]), config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bootstrap config: %w", err)
	}
	return config, nil
}

type ClusterRole string

const (
	ClusterRoleControlPlane ClusterRole = "control-plane"
	ClusterRoleWorker       ClusterRole = "worker"
	// The role of a node is unknown if it has not yet joined a cluster,
	// currently joining or is about to leave.
	ClusterRoleUnknown ClusterRole = "unknown"
)

// DatastoreRole as provided by dqlite
type DatastoreRole string

const (
	DatastoreRoleVoter   DatastoreRole = "voter"
	DatastoreRoleStandBy DatastoreRole = "stand-by"
	DatastoreRoleSpare   DatastoreRole = "spare"
	DatastoreRolePending DatastoreRole = "PENDING"
	DatastoreRoleUnknown DatastoreRole = "unknown"
)

// NodeStatus holds information about a node in the k8s cluster.
type NodeStatus struct {
	// Name is the name for this cluster member that was when joining the cluster.
	// This is typically the hostname of the node.
	Name string `json:"name,omitempty"`
	// Address is the IP address of the node.
	Address string `json:"address,omitempty"`
	// ClusterRole is the role that the node has within the k8s cluster.
	ClusterRole ClusterRole `json:"cluster-role,omitempty"`
	// DatastoreRole is the role that the node has within the datastore cluster.
	// Only applicable for control-plane nodes, empty for workers.
	DatastoreRole DatastoreRole `json:"datastore-role,omitempty"`
}

// ClusterStatus holds information about the cluster, e.g. its current members
type ClusterStatus struct {
	// Ready is true if at least one node in the cluster is in READY state.
	Ready      bool         `json:"ready,omitempty"`
	Members    []NodeStatus `json:"members,omitempty"`
	Components []Component  `json:"components,omitempty"`
}

// HaClusterFormed returns true if the cluster is in high-availability mode (more than two voter nodes).
func (c ClusterStatus) HaClusterFormed() bool {
	voters := 0
	for _, member := range c.Members {
		if member.DatastoreRole == DatastoreRoleVoter {
			voters++
		}
	}
	return voters > 2
}

func (c ClusterStatus) String() string {
	result := strings.Builder{}

	if c.Ready {
		result.WriteString("k8s is ready.")
	} else {
		result.WriteString("k8s is not ready.")
	}
	result.WriteString("\n")

	result.WriteString("high-availability: ")
	if c.HaClusterFormed() {
		result.WriteString("yes")
	} else {
		result.WriteString("no")
	}
	result.WriteString("\n\n")
	result.WriteString("control-plane nodes:\n")
	for _, member := range c.Members {
		// There is not much that we can do if the hostport is wrong.
		// Thus, ignore the error and just display an empty IP field.
		apiServerIp, _, _ := net.SplitHostPort(member.Address)
		result.WriteString(fmt.Sprintf("  %s: %s\n", member.Name, apiServerIp))
	}
	result.WriteString("\n")

	result.WriteString("components:\n")
	for _, component := range c.Components {
		result.WriteString(fmt.Sprintf("  %-10s %s\n", component.Name, component.Status))
	}

	return result.String()
}
