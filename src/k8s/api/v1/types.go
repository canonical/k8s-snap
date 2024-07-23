package apiv1

import (
	"fmt"
	"strings"
	"time"
)

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

// FeatureStatus encapsulates the deployment status of a feature.
type FeatureStatus struct {
	// Enabled shows whether or not the deployment of manifests for a status was successful.
	Enabled bool
	// Message contains information about the status of a feature. It is only supposed to be human readable and informative and should not be programmatically parsed.
	Message string
	// Version shows the version of the deployed feature.
	Version string
	// UpdatedAt shows when the last update was done.
	UpdatedAt time.Time
}

func (f FeatureStatus) String() string {
	if f.Message != "" {
		return f.Message
	}
	if f.Enabled {
		return "enabled"
	}
	return "disabled"
}

type Datastore struct {
	Type    string   `json:"type,omitempty"`
	Servers []string `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// ClusterStatus holds information about the cluster, e.g. its current members
type ClusterStatus struct {
	// Ready is true if at least one node in the cluster is in READY state.
	Ready     bool                    `json:"ready,omitempty"`
	Members   []NodeStatus            `json:"members,omitempty"`
	Config    UserFacingClusterConfig `json:"config,omitempty"`
	Datastore Datastore               `json:"datastore,omitempty"`

	DNS           FeatureStatus `json:"dns,omitempty"`
	Network       FeatureStatus `json:"network,omitempty"`
	LoadBalancer  FeatureStatus `json:"load-balancer,omitempty"`
	Ingress       FeatureStatus `json:"ingress,omitempty"`
	Gateway       FeatureStatus `json:"gateway,omitempty"`
	MetricsServer FeatureStatus `json:"metrics-server,omitempty"`
	LocalStorage  FeatureStatus `json:"local-storage,omitempty"`
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

// TICS -COV_GO_SUPPRESSED_ERROR
// we are just formatting the output for the k8s status command, it is ok to ignore failures from result.WriteString()

// TODO: Print k8s version. However, multiple nodes can run different version, so we would need to query all nodes.
func (c ClusterStatus) String() string {
	result := strings.Builder{}

	// Status
	if c.Ready {
		result.WriteString(fmt.Sprintf("%-25s %s", "cluster status:", "ready"))
	} else {
		result.WriteString(fmt.Sprintf("%-25s %s", "cluster status:", "not ready"))
	}
	result.WriteString("\n")

	// Control Plane Nodes
	result.WriteString(fmt.Sprintf("%-25s ", "control plane nodes:"))
	if len(c.Members) > 0 {
		members := make([]string, 0, len(c.Members))
		for _, m := range c.Members {
			members = append(members, fmt.Sprintf("%s (%s)", m.Address, m.DatastoreRole))
		}
		result.WriteString(strings.Join(members, ", "))
	} else {
		result.WriteString("none")
	}
	result.WriteString("\n")

	// High availability
	result.WriteString(fmt.Sprintf("%-25s ", "high availability:"))
	if c.HaClusterFormed() {
		result.WriteString("yes")
	} else {
		result.WriteString("no")
	}
	result.WriteString("\n")

	// Datastore
	// TODO: how to understand if the ds is running or not?
	if c.Datastore.Type != "" {
		result.WriteString(fmt.Sprintf("%-25s %s\n", "datastore:", c.Datastore.Type))
	} else {
		result.WriteString(fmt.Sprintf("%-25s %s\n", "datastore:", "disabled"))
	}

	result.WriteString(fmt.Sprintf("%-25s %s\n", "network:", c.Network))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "dns:", c.DNS))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "ingress:", c.Ingress))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "load-balancer:", c.LoadBalancer))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "local-storage:", c.LocalStorage))
	result.WriteString(fmt.Sprintf("%-25s %s\n", "gateway", c.Gateway))

	return result.String()
}

// TICS +COV_GO_SUPPRESSED_ERROR
