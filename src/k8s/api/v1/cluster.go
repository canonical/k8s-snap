package v1

import (
	"fmt"
	"strings"
)

// GetClusterStatusRequest is used to request the current status of the cluster.
type GetClusterStatusRequest struct{}

// GetClusterStatusResponse is the response for "GET 1.0/k8sd/cluster".
type GetClusterStatusResponse struct {
	ClusterStatus ClusterStatus `json:"status"`
}

// InitClusterRequest is used to initialize a k8s cluster.
type InitClusterRequest struct{}

// InitClusterResponse is the response for "POST 1.0/k8sd/cluster".
type InitClusterResponse struct{}

// GetKubeConfigRequest is used to ask for the admin kubeconfig
type GetKubeConfigRequest struct{}

// GetKubeConfigResponse is the response for "GET 1.0/k8sd/cluster/config".
type GetKubeConfigResponse struct {
	KubeConfig string `json:"kubeconfig"`
}

// JoinClusterRequest is used to request the configuration to join the k8s cluster.
type JoinClusterRequest struct {
	Token string `json:"token"`
}

// JoinClusterResponse is the response for "POST 1.0/k8sd/cluster/join".
type JoinClusterResponse struct {
	// ExtraServiceArgs overwrites the configuration of
	// kube services on the joining node so that they can connect with
	// the master node services.
	// TODO: use named arguments (e.g. ExtraKubeletArgs) instead.
	ExtraServiceArgs ExtraServiceArgs `json:"extraServiceArgs"`
}

// ExtraServiceArgs specify k8s service arguments that should be overwritten.
//
//	ServiceName:{
//	 "--argument": "value"
//	}
type ExtraServiceArgs map[string]map[string]string

// ClusterMember holds information about a node in the k8s cluster.
type ClusterMember struct {
	Name        string `mapstructure:"name,omitempty"`
	Address     string `mapstructure:"address,omitempty"`
	Role        string `mapstructure:"role,omitempty"`
	Fingerprint string `mapstructure:"fingerprint,omitempty"`
	Status      string `mapstructure:"status,omitempty"`
}

// ClusterStatus holds information about the cluster, e.g. its current members
type ClusterStatus struct {
	// Ready is true if at least one node in the cluster is in READY state.
	Ready      bool            `mapstructure:"ready,omitempty"`
	Members    []ClusterMember `mapstructure:"members,omitempty"`
	Components []Component     `mapstructure:"components,omitempty"`
}

// HaClusterFormed returns true if the cluster is in high-availability mode (more than two voter nodes).
func (c ClusterStatus) HaClusterFormed() bool {
	voters := 0
	for _, member := range c.Members {
		if member.Role == "voter" {
			voters++
		}
	}
	return voters > 2
}

func (c ClusterStatus) String() string {
	result := strings.Builder{}

	if c.Ready {
		result.WriteString("k8s is running")
	} else {
		result.WriteString("k8s is not running.\n")
		return result.String()
	}
	result.WriteString("\n")

	result.WriteString("high-availability: ")
	if c.HaClusterFormed() {
		result.WriteString("yes")
	} else {
		result.WriteString("no")
	}
	result.WriteString("\n\n")

	result.WriteString("components:\n")
	for _, component := range c.Components {
		result.WriteString(fmt.Sprintf("  %s: %s\n", component.Name, component.Status))
	}

	return result.String()
}
