package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestSetDefaults(t *testing.T) {
	g := NewWithT(t)
	clusterConfig := types.ClusterConfig{}

	// Set defaults
	expectedConfig := types.ClusterConfig{
		Network: types.Network{
			Enabled:     vals.Pointer(false),
			PodCIDR:     vals.Pointer("10.1.0.0/16"),
			ServiceCIDR: vals.Pointer("10.152.183.0/24"),
		},
		APIServer: types.APIServer{
			SecurePort:        vals.Pointer(6443),
			AuthorizationMode: vals.Pointer("Node,RBAC"),
		},
		Datastore: types.Datastore{
			Type:          vals.Pointer("k8s-dqlite"),
			K8sDqlitePort: vals.Pointer(9000),
		},
		Kubelet: types.Kubelet{
			ClusterDomain: vals.Pointer("cluster.local"),
		},
		DNS: types.DNS{
			Enabled:             vals.Pointer(false),
			UpstreamNameservers: vals.Pointer([]string{"/etc/resolv.conf"}),
		},
		LocalStorage: types.LocalStorage{
			Enabled:       vals.Pointer(false),
			LocalPath:     vals.Pointer("/var/snap/k8s/common/rawfile-storage"),
			ReclaimPolicy: vals.Pointer("Delete"),
			Default:       vals.Pointer(true),
		},
		LoadBalancer: types.LoadBalancer{
			Enabled:        vals.Pointer(false),
			CIDRs:          vals.Pointer([]string{}),
			L2Mode:         vals.Pointer(false),
			L2Interfaces:   vals.Pointer([]string{}),
			BGPMode:        vals.Pointer(false),
			BGPLocalASN:    vals.Pointer(0),
			BGPPeerAddress: vals.Pointer(""),
			BGPPeerASN:     vals.Pointer(0),
			BGPPeerPort:    vals.Pointer(0),
		},
		MetricsServer: types.MetricsServer{
			Enabled: vals.Pointer(true),
		},
		Gateway: types.Gateway{
			Enabled: vals.Pointer(false),
		},
		Ingress: types.Ingress{
			Enabled:             vals.Pointer(false),
			DefaultTLSSecret:    vals.Pointer(""),
			EnableProxyProtocol: vals.Pointer(false),
		},
	}

	clusterConfig.SetDefaults()
	g.Expect(clusterConfig).To(Equal(expectedConfig))
}
