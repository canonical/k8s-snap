package types_test

import (
	"testing"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestClusterConfigFromBootstrapConfig(t *testing.T) {
	for _, tc := range []struct {
		name         string
		bootstrap    apiv1.BootstrapConfig
		expectConfig types.ClusterConfig
	}{
		{
			name: "Nil",
			expectConfig: types.ClusterConfig{
				APIServer: types.APIServer{
					AuthorizationMode: vals.Pointer("Node,RBAC"),
				},
				Datastore: types.Datastore{
					Type: vals.Pointer("k8s-dqlite"),
				},
			},
		},
		{
			name: "DisableRBAC",
			bootstrap: apiv1.BootstrapConfig{
				DisableRBAC: vals.Pointer(true),
			},
			expectConfig: types.ClusterConfig{
				APIServer: types.APIServer{
					AuthorizationMode: vals.Pointer("AlwaysAllow"),
				},
				Datastore: types.Datastore{
					Type: vals.Pointer("k8s-dqlite"),
				},
			},
		},
		{
			name: "K8sDqliteDefault",
			bootstrap: apiv1.BootstrapConfig{
				DatastoreType: vals.Pointer(""),
			},
			expectConfig: types.ClusterConfig{
				APIServer: types.APIServer{
					AuthorizationMode: vals.Pointer("Node,RBAC"),
				},
				Datastore: types.Datastore{
					Type: vals.Pointer("k8s-dqlite"),
				},
			},
		},
		{
			name: "ExternalDatastore",
			bootstrap: apiv1.BootstrapConfig{
				DatastoreType:       vals.Pointer("external"),
				DatastoreServers:    []string{"https://10.0.0.1:2379", "https://10.0.0.2:2379"},
				DatastoreCACert:     vals.Pointer("CA DATA"),
				DatastoreClientCert: vals.Pointer("CERT DATA"),
				DatastoreClientKey:  vals.Pointer("KEY DATA"),
			},
			expectConfig: types.ClusterConfig{
				APIServer: types.APIServer{
					AuthorizationMode: vals.Pointer("Node,RBAC"),
				},
				Datastore: types.Datastore{
					Type:               vals.Pointer("external"),
					ExternalURL:        vals.Pointer("https://10.0.0.1:2379,https://10.0.0.2:2379"),
					ExternalCACert:     vals.Pointer("CA DATA"),
					ExternalClientCert: vals.Pointer("CERT DATA"),
					ExternalClientKey:  vals.Pointer("KEY DATA"),
				},
			},
		},
		{
			name: "Full",
			bootstrap: apiv1.BootstrapConfig{
				ClusterConfig: apiv1.UserFacingClusterConfig{
					Network: apiv1.NetworkConfig{
						Enabled: vals.Pointer(true),
					},
					DNS: apiv1.DNSConfig{
						Enabled:       vals.Pointer(true),
						ClusterDomain: vals.Pointer("cluster.local"),
					},
					Ingress: apiv1.IngressConfig{
						Enabled: vals.Pointer(true),
					},
					LoadBalancer: apiv1.LoadBalancerConfig{
						Enabled: vals.Pointer(true),
						L2Mode:  vals.Pointer(true),
						CIDRs:   vals.Pointer([]string{"10.0.0.0/24", "10.1.0.10-10.1.0.20"}),
					},
					LocalStorage: apiv1.LocalStorageConfig{
						Enabled:   vals.Pointer(true),
						LocalPath: vals.Pointer("/storage/path"),
						Default:   vals.Pointer(false),
					},
					Gateway: apiv1.GatewayConfig{
						Enabled: vals.Pointer(true),
					},
					MetricsServer: apiv1.MetricsServerConfig{
						Enabled: vals.Pointer(true),
					},
					CloudProvider: vals.Pointer("external"),
				},
				PodCIDR:       vals.Pointer("10.100.0.0/16"),
				ServiceCIDR:   vals.Pointer("10.200.0.0/16"),
				DisableRBAC:   vals.Pointer(false),
				SecurePort:    vals.Pointer(6443),
				K8sDqlitePort: vals.Pointer(9090),
				DatastoreType: vals.Pointer("k8s-dqlite"),
				ExtraSANs:     []string{"custom.kubernetes"},
			},
			expectConfig: types.ClusterConfig{
				Datastore: types.Datastore{
					Type:          vals.Pointer("k8s-dqlite"),
					K8sDqlitePort: vals.Pointer(9090),
				},
				APIServer: types.APIServer{
					SecurePort:        vals.Pointer(6443),
					AuthorizationMode: vals.Pointer("Node,RBAC"),
				},
				Kubelet: types.Kubelet{
					ClusterDomain: vals.Pointer("cluster.local"),
					CloudProvider: vals.Pointer("external"),
				},
				Network: types.Network{
					Enabled:     vals.Pointer(true),
					PodCIDR:     vals.Pointer("10.100.0.0/16"),
					ServiceCIDR: vals.Pointer("10.200.0.0/16"),
				},
				DNS: types.DNS{
					Enabled: vals.Pointer(true),
				},
				Ingress: types.Ingress{
					Enabled: vals.Pointer(true),
				},
				LoadBalancer: types.LoadBalancer{
					Enabled:  vals.Pointer(true),
					L2Mode:   vals.Pointer(true),
					CIDRs:    vals.Pointer([]string{"10.0.0.0/24"}),
					IPRanges: vals.Pointer([]types.LoadBalancer_IPRange{{Start: "10.1.0.10", Stop: "10.1.0.20"}}),
				},
				LocalStorage: types.LocalStorage{
					Enabled:   vals.Pointer(true),
					LocalPath: vals.Pointer("/storage/path"),
					Default:   vals.Pointer(false),
				},
				Gateway: types.Gateway{
					Enabled: vals.Pointer(true),
				},
				MetricsServer: types.MetricsServer{
					Enabled: vals.Pointer(true),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			config, err := types.ClusterConfigFromBootstrapConfig(tc.bootstrap)
			g.Expect(err).To(BeNil())
			g.Expect(config).To(Equal(tc.expectConfig))
		})
	}

	t.Run("Invalid", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			bootstrap apiv1.BootstrapConfig
		}{
			{
				name: "K8sDqliteWithExternalServers",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType:    vals.Pointer(""),
					DatastoreServers: []string{"http://10.0.0.1:2379"},
				},
			},
			{
				name: "K8sDqliteWithExternalCA",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType:   vals.Pointer(""),
					DatastoreCACert: vals.Pointer("CA DATA"),
				},
			},
			{
				name: "K8sDqliteWithExternalClientCert",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType:       vals.Pointer(""),
					DatastoreClientCert: vals.Pointer("CERT DATA"),
				},
			},
			{
				name: "K8sDqliteWithExternalClientKey",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType:      vals.Pointer(""),
					DatastoreClientKey: vals.Pointer("KEY DATA"),
				},
			},
			{
				name: "ExternalWithK8sDqlitePort",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType:    vals.Pointer("external"),
					DatastoreServers: []string{"http://10.0.0.1:2379"},
					K8sDqlitePort:    vals.Pointer(18080),
				},
			},
			{
				name: "ExternalWithoutServers",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType: vals.Pointer("external"),
				},
			},
			{
				name: "UnsupportedDatastore",
				bootstrap: apiv1.BootstrapConfig{
					DatastoreType: vals.Pointer("unknown"),
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				config, err := types.ClusterConfigFromBootstrapConfig(tc.bootstrap)
				g.Expect(config).To(BeZero())
				g.Expect(err).To(HaveOccurred())
			})
		}

	})
}
