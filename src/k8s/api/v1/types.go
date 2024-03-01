package v1

import (
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/utils/vals"
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
	b.Components = []string{"dns", "metrics-server", "network"}
	b.ClusterCIDR = "10.1.0.0/16"
	b.EnableRBAC = vals.Pointer(true)
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
	Ready   bool                    `json:"ready,omitempty"`
	Members []NodeStatus            `json:"members,omitempty"`
	Config  UserFacingClusterConfig `json:"config,omitempty"`
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

// TODO: Print k8s version. However, multiple nodes can run different version, so we would need to query all nodes.
func (c ClusterStatus) String() string {
	result := strings.Builder{}

	if c.Ready {
		result.WriteString("status: ready")
	} else {
		result.WriteString("status: not ready")
	}
	result.WriteString("\n")

	result.WriteString("high-availability: ")
	if c.HaClusterFormed() {
		result.WriteString("yes")
	} else {
		result.WriteString("no")
	}
	result.WriteString("\n")
	result.WriteString("datastore:\n")

	voters := make([]NodeStatus, 0, len(c.Members))
	standBys := make([]NodeStatus, 0, len(c.Members))
	spares := make([]NodeStatus, 0, len(c.Members))
	for _, node := range c.Members {
		switch node.DatastoreRole {
		case DatastoreRoleVoter:
			voters = append(voters, node)
		case DatastoreRoleStandBy:
			standBys = append(standBys, node)
		case DatastoreRoleSpare:
			spares = append(spares, node)
		}
	}
	if len(voters) > 0 {
		result.WriteString(fmt.Sprintf("  voter-nodes:\n"))
		for _, voter := range voters {
			result.WriteString(fmt.Sprintf("    - %s\n", voter.Address))
		}
	} else {
		result.WriteString(fmt.Sprintf("  voter-nodes: none\n"))
	}
	if len(standBys) > 0 {
		result.WriteString(fmt.Sprintf("  standby-nodes:\n"))
		for _, standBy := range standBys {
			result.WriteString(fmt.Sprintf("    - %s\n", standBy.Address))
		}
	} else {
		result.WriteString(fmt.Sprintf("  standy-nodes: none\n"))
	}
	if len(spares) > 0 {
		result.WriteString(fmt.Sprintf("  spare-nodes:\n"))
		for _, spare := range spares {
			result.WriteString(fmt.Sprintf("    - %s\n", spare.Address))
		}
	} else {
		result.WriteString(fmt.Sprintf("  spare-nodes: none\n"))
	}
	result.WriteString("\n")

	printedConfig := UserFacingClusterConfig{}
	if c.Config.Network.Enabled != nil && *c.Config.Network.Enabled {
		printedConfig.Network = c.Config.Network
	}
	if c.Config.DNS.Enabled != nil && *c.Config.DNS.Enabled {
		printedConfig.DNS = c.Config.DNS
	}
	if c.Config.Ingress.Enabled != nil && *c.Config.Ingress.Enabled {
		printedConfig.Ingress = c.Config.Ingress
	}
	if c.Config.LoadBalancer.Enabled != nil && *c.Config.LoadBalancer.Enabled {
		printedConfig.LoadBalancer = c.Config.LoadBalancer
	}
	if c.Config.LocalStorage.Enabled != nil && *c.Config.LocalStorage.Enabled {
		printedConfig.LocalStorage = c.Config.LocalStorage
	}
	if c.Config.Gateway.Enabled != nil && *c.Config.Gateway.Enabled {
		printedConfig.Gateway = c.Config.Gateway
	}
	if c.Config.MetricsServer.Enabled != nil && *c.Config.MetricsServer.Enabled {
		printedConfig.MetricsServer = c.Config.MetricsServer
	}

	b, _ := yaml.Marshal(printedConfig)
	result.WriteString(string(b))
	result.WriteString("\n")

	return result.String()
}
