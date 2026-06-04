package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
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
			Type:          utils.Pointer("etcd"),
			K8sDqlitePort: utils.Pointer(9000),
			EtcdPort:      utils.Pointer(2379),
			EtcdPeerPort:  utils.Pointer(2380),
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
			L2Mode:         utils.Pointer(true),
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

func TestControlPlaneEndpointDefaults(t *testing.T) {
	for _, tc := range []struct {
		name          string
		config        types.ClusterConfig
		expectHost    string
		expectPort    int
		expectBackend string
	}{
		{
			name:   "NoHost/NotDefaulted",
			config: types.ClusterConfig{},
		},
		{
			name: "HostSet/DefaultsPortAndBackend",
			config: types.ClusterConfig{
				ControlPlaneEndpoint: types.ControlPlaneEndpoint{Host: utils.Pointer("10.0.0.250")},
			},
			expectHost:    "10.0.0.250",
			expectPort:    6443,
			expectBackend: "external",
		},
		{
			name: "HostSet/RespectsExplicitPortAndBackend",
			config: types.ClusterConfig{
				ControlPlaneEndpoint: types.ControlPlaneEndpoint{
					Host:    utils.Pointer("api.example.com"),
					Port:    utils.Pointer(443),
					Backend: utils.Pointer("service"),
				},
			},
			expectHost:    "api.example.com",
			expectPort:    443,
			expectBackend: "service",
		},
		{
			name: "HostSet/PortDefaultsToCustomSecurePort",
			config: types.ClusterConfig{
				APIServer:            types.APIServer{SecurePort: utils.Pointer(7443)},
				ControlPlaneEndpoint: types.ControlPlaneEndpoint{Host: utils.Pointer("10.0.0.250")},
			},
			expectHost:    "10.0.0.250",
			expectPort:    7443,
			expectBackend: "external",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			tc.config.SetDefaults()
			g.Expect(tc.config.ControlPlaneEndpoint.GetHost()).To(Equal(tc.expectHost))
			g.Expect(tc.config.ControlPlaneEndpoint.GetPort()).To(Equal(tc.expectPort))
			g.Expect(tc.config.ControlPlaneEndpoint.GetBackend()).To(Equal(tc.expectBackend))
		})
	}
}
