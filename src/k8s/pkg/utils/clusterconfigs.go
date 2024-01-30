package utils

import (
	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database/clusterconfigs"
)

// ConvertBootstrapToClusterConfig extracts the cluster config parts from the BootstrapConfig
// and maps them to a ClusterConfig.
func ConvertBootstrapToClusterConfig(b *apiv1.BootstrapConfig) clusterconfigs.ClusterConfig {
	return clusterconfigs.ClusterConfig{
		Cluster: clusterconfigs.Cluster{
			CIDR: b.ClusterCIDR,
		},
	}
}
