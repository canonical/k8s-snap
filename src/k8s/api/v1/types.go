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
	// ServiceCIDR is the CIDR of the cluster services.
	ServiceCIDR string `yaml:"service-cidr"`
	// EnableRBAC determines if RBAC will be enabled; *bool to know true/false/unset.
	EnableRBAC          *bool    `yaml:"enable-rbac"`
	K8sDqlitePort       int      `yaml:"k8s-dqlite-port"`
	Datastore           string   `yaml:"datastore"`
	DatastoreURL        string   `yaml:"datastore-url,omitempty"`
	DatastoreCACert     string   `yaml:"datastore-ca-crt,omitempty"`
	DatastoreClientCert string   `yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey  string   `yaml:"datastore-client-key,omitempty"`
	ExtraSANs           []string `yaml:"extrasans,omitempty"`
}

// SetDefaults sets the fields to default values.
func (b *BootstrapConfig) SetDefaults() {
	b.Components = []string{"dns", "metrics-server", "network", "gateway"}
	b.ClusterCIDR = "10.1.0.0/16"
	b.ServiceCIDR = "10.152.183.0/24"
	b.EnableRBAC = vals.Pointer(true)
	b.K8sDqlitePort = 9000
	b.Datastore = "k8s-dqlite"
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

type Datastore struct {
	Type        string `json:"type,omitempty"`
	ExternalURL string `json:"external-url,omitempty" yaml:"external-url,omitempty"`
}

// ClusterStatus holds information about the cluster, e.g. its current members
type ClusterStatus struct {
	// Ready is true if at least one node in the cluster is in READY state.
	Ready     bool                    `json:"ready,omitempty"`
	Members   []NodeStatus            `json:"members,omitempty"`
	Config    UserFacingClusterConfig `json:"config,omitempty"`
	Datastore Datastore               `json:"datastore,omitempty"`
}

// haClusterFormed returns true if the cluster is in high-availability mode (more than two voter nodes).
func (c ClusterStatus) haClusterFormed() bool {
	voters := 0
	for _, member := range c.Members {
		if member.DatastoreRole == DatastoreRoleVoter {
			voters++
		}
	}
	return voters > 2
}

func (c ClusterStatus) datastoreToString() string {
	result := strings.Builder{}

	// Datastore
	if c.Datastore.Type != "" {
		result.WriteString(fmt.Sprintf("  type: %s\n", c.Datastore.Type))
		// Datastore URL for external only
		if c.Datastore.Type == "external" {
			if c.Datastore.ExternalURL != "" {
				result.WriteString(fmt.Sprintf("  url: %s\n", c.Datastore.ExternalURL))
			}
			return result.String()
		}
	}

	// Datastore roles for dqlite
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
		result.WriteString("  voter-nodes:\n")
		for _, voter := range voters {
			result.WriteString(fmt.Sprintf("    - %s\n", voter.Address))
		}
	} else {
		result.WriteString("  voter-nodes: none\n")
	}
	if len(standBys) > 0 {
		result.WriteString("  standby-nodes:\n")
		for _, standBy := range standBys {
			result.WriteString(fmt.Sprintf("    - %s\n", standBy.Address))
		}
	} else {
		result.WriteString("  standby-nodes: none\n")
	}
	if len(spares) > 0 {
		result.WriteString("  spare-nodes:\n")
		for _, spare := range spares {
			result.WriteString(fmt.Sprintf("    - %s\n", spare.Address))
		}
	} else {
		result.WriteString("  spare-nodes: none\n")
	}

	return result.String()
}

// TODO: Print k8s version. However, multiple nodes can run different version, so we would need to query all nodes.
func (c ClusterStatus) String() string {
	result := strings.Builder{}

	// Status
	if c.Ready {
		result.WriteString("status: ready")
	} else {
		result.WriteString("status: not ready")
	}
	result.WriteString("\n")

	// High availability
	result.WriteString("high-availability: ")
	if c.haClusterFormed() {
		result.WriteString("yes")
	} else {
		result.WriteString("no")
	}

	// Datastore
	result.WriteString("\n")
	result.WriteString("datastore:\n")
	result.WriteString(c.datastoreToString())

	// Config
	var emptyConfig UserFacingClusterConfig
	if c.Config != emptyConfig {
		b, _ := yaml.Marshal(c.Config)
		result.WriteString(string(b))
	}
	return result.String()
}
