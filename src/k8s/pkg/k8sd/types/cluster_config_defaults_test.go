package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"

	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestSetDefaults(t *testing.T) {
	g := NewWithT(t)
	clusterConfig := types.ClusterConfig{}

	// Set defaults
	expectedConfig := types.ClusterConfig{
		Network: types.Network{
			Enabled:     utils.Pointer(false),
			PodCIDR:     utils.Pointer("10.1.0.0/16"),
			ServiceCIDR: utils.Pointer("10.152.183.0/24"),
		},
		APIServer: types.APIServer{
			SecurePort:        utils.Pointer(6443),
			AuthorizationMode: utils.Pointer("Node,RBAC"),
		},
		Datastore: types.Datastore{
			Type:          utils.Pointer("k8s-dqlite"),
			K8sDqlitePort: utils.Pointer(9000),
		},
		K8sd: types.K8sd{
			ShouldRemoveK8sNode: utils.Pointer(true),
		},
		Kubelet: types.Kubelet{
			ClusterDomain: utils.Pointer("cluster.local"),
		},
		DNS: types.DNS{
			Enabled:             utils.Pointer(false),
			UpstreamNameservers: utils.Pointer([]string{"/etc/resolv.conf"}),
		},
		LocalStorage: types.LocalStorage{
			Enabled:       utils.Pointer(false),
			LocalPath:     utils.Pointer("/var/snap/k8s/common/rawfile-storage"),
			ReclaimPolicy: utils.Pointer("Delete"),
			Default:       utils.Pointer(true),
		},
		LoadBalancer: types.LoadBalancer{
			Enabled:        utils.Pointer(false),
			CIDRs:          utils.Pointer([]string{}),
			L2Mode:         utils.Pointer(false),
			L2Interfaces:   utils.Pointer([]string{}),
			BGPMode:        utils.Pointer(false),
			BGPLocalASN:    utils.Pointer(0),
			BGPPeerAddress: utils.Pointer(""),
			BGPPeerASN:     utils.Pointer(0),
			BGPPeerPort:    utils.Pointer(0),
		},
		MetricsServer: types.MetricsServer{
			Enabled: utils.Pointer(true),
		},
		Gateway: types.Gateway{
			Enabled: utils.Pointer(false),
		},
		Ingress: types.Ingress{
			Enabled:             utils.Pointer(false),
			DefaultTLSSecret:    utils.Pointer(""),
			EnableProxyProtocol: utils.Pointer(false),
		},
	}

	clusterConfig.SetDefaults()
	g.Expect(clusterConfig).To(Equal(expectedConfig))
}
