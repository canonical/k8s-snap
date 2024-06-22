package types_test

import (
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/utils"

	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

type mergeClusterConfigTestCase struct {
	name         string
	old          types.ClusterConfig
	new          types.ClusterConfig
	expectResult types.ClusterConfig
	expectErr    bool
}

func generateMergeClusterConfigTestCases[T any](field string, changeAllowed bool, val1 T, val2 T, update func(*types.ClusterConfig, any)) []mergeClusterConfigTestCase {
	var cfgNil, cfgZero, cfgOne, cfgTwo types.ClusterConfig
	var zero T

	// defaults for validation
	for _, cfg := range []*types.ClusterConfig{&cfgNil, &cfgZero, &cfgOne, &cfgTwo} {
		cfg.Network.PodCIDR = utils.Pointer("10.1.0.0/16")
		cfg.Network.ServiceCIDR = utils.Pointer("10.152.183.0/24")
	}

	update(&cfgZero, zero)
	update(&cfgOne, val1)
	update(&cfgTwo, val2)

	return []mergeClusterConfigTestCase{
		{
			name:         fmt.Sprintf("%s/Empty", field),
			old:          cfgNil,
			new:          cfgNil,
			expectResult: cfgNil,
		},
		{
			name:         fmt.Sprintf("%s/Set", field),
			new:          cfgOne,
			expectResult: cfgOne,
			expectErr:    false,
		},
		{
			name:         fmt.Sprintf("%s/Keep", field),
			old:          cfgOne,
			new:          cfgNil,
			expectResult: cfgOne,
		},
		{
			name:         fmt.Sprintf("%s/Update", field),
			old:          cfgOne,
			new:          cfgTwo,
			expectResult: cfgTwo,
			expectErr:    !changeAllowed,
		},
		{
			name:         fmt.Sprintf("%s/Unset", field),
			old:          cfgOne,
			new:          cfgZero,
			expectResult: cfgZero,
			expectErr:    !changeAllowed,
		},
	}
}

func TestMergeClusterConfig(t *testing.T) {
	for _, tcs := range [][]mergeClusterConfigTestCase{
		generateMergeClusterConfigTestCases("Certificates/CACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.CACert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/CAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.CAKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/ClientCACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.ClientCACert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/ClientCAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.ClientCAKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/FrontProxyCACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.FrontProxyCACert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/FrontProxyCAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.FrontProxyCAKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/ServiceAccountKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.ServiceAccountKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/APIServerKubeletClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) {
			c.Certificates.APIServerKubeletClientCert = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Certificates/APIServerKubeletClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) {
			c.Certificates.APIServerKubeletClientKey = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Certificates/AdminClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.AdminClientCert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/AdminClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.AdminClientKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/K8sdPublicKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.K8sdPublicKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/K8sdPrivateKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.K8sdPrivateKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/Type", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.Type = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqliteCert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.K8sDqliteCert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqliteKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.K8sDqliteKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqlitePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.Datastore.K8sDqlitePort = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalServers", true, []string{"localhost:123"}, []string{"localhost:123"}, func(c *types.ClusterConfig, v any) { c.Datastore.ExternalServers = utils.Pointer(v.([]string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalCACert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalCACert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalClientCert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalClientKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/EtcdCACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.EtcdCACert = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/EtcdCAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.EtcdCAKey = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/EtcdAPIServerClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) {
			c.Datastore.EtcdAPIServerClientCert = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Datastore/EtcdAPIServerClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) {
			c.Datastore.EtcdAPIServerClientKey = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Datastore/EtcdPort", false, 2379, 12379, func(c *types.ClusterConfig, v any) { c.Datastore.EtcdPort = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Datastore/EtcdPeerPort", false, 2380, 12380, func(c *types.ClusterConfig, v any) { c.Datastore.EtcdPeerPort = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Network/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Network.Enabled = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Network/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Network.Enabled = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Network/PodCIDR", false, "10.1.0.0/16", "10.2.0.0/16", func(c *types.ClusterConfig, v any) { c.Network.PodCIDR = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/ServiceCIDR", false, "10.152.183.0/24", "10.152.184.0/24", func(c *types.ClusterConfig, v any) { c.Network.ServiceCIDR = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("APIServer/SecurePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.APIServer.SecurePort = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("APIServer/AuthorizationMode", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.AuthorizationMode = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/CloudProvider", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.CloudProvider = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDNS/AllowChange", true, "1.1.1.1", "2.2.2.2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDNS = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDNS/PreventChangeIfDNSEnabled", false, "1.1.1.1", "2.2.2.2", func(c *types.ClusterConfig, v any) {
			c.DNS.Enabled = utils.Pointer(true)
			c.Kubelet.ClusterDNS = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDomain", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDomain = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("DNS/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.DNS.Enabled = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("DNS/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.DNS.Enabled = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("DNS/UpstreamNameservers", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.DNS.UpstreamNameservers = utils.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Ingress/Enable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = utils.Pointer(true)
			c.Ingress.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Ingress/Disable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = utils.Pointer(true)
			c.Ingress.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Ingress/DefaultTLSSecret", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Ingress.DefaultTLSSecret = utils.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Ingress/EnableProxyProtocol/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Ingress.EnableProxyProtocol = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Ingress/EnableProxyProtocol/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Ingress.EnableProxyProtocol = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Gateway/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = utils.Pointer(true)
			c.Gateway.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Gateway/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = utils.Pointer(true)
			c.Gateway.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = utils.Pointer(true)
			c.LoadBalancer.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = utils.Pointer(true)
			c.LoadBalancer.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/CIDRs", true, []string{"172.16.101.0/24"}, []string{"172.16.102.0/24"}, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.CIDRs = utils.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/IPRanges", true, []types.LoadBalancer_IPRange{{Start: "10.0.0.10", Stop: "10.0.0.20"}}, []types.LoadBalancer_IPRange{{Start: "10.1.0.10", Stop: "10.1.0.20"}}, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.IPRanges = utils.Pointer(v.([]types.LoadBalancer_IPRange))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/L2Mode/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.LoadBalancer.L2Mode = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/L2Mode/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.LoadBalancer.L2Mode = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/L2Interfaces", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.L2Interfaces = utils.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPMode/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.BGPMode = utils.Pointer(v.(bool))
			c.LoadBalancer.BGPLocalASN = utils.Pointer(100)
			c.LoadBalancer.BGPPeerAddress = utils.Pointer("10.10.0.0/16")
			c.LoadBalancer.BGPPeerASN = utils.Pointer(101)
			c.LoadBalancer.BGPPeerPort = utils.Pointer(10010)
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPMode/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.BGPMode = utils.Pointer(v.(bool))
			c.LoadBalancer.BGPLocalASN = utils.Pointer(100)
			c.LoadBalancer.BGPPeerAddress = utils.Pointer("10.10.0.0/16")
			c.LoadBalancer.BGPPeerASN = utils.Pointer(101)
			c.LoadBalancer.BGPPeerPort = utils.Pointer(10010)
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPLocalASN", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.LoadBalancer.BGPLocalASN = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPPeerAddress", true, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.BGPPeerAddress = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPPeerASN", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.LoadBalancer.BGPPeerASN = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPPeerPort", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.LoadBalancer.BGPPeerPort = utils.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("LocalStorage/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.LocalStorage.LocalPath = utils.Pointer("path")
			c.LocalStorage.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.LocalStorage.LocalPath = utils.Pointer("path")
			c.LocalStorage.Enabled = utils.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/LocalPath/AllowChange", true, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.LocalPath = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/LocalPath/PreventChangeIfEnabled", false, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.Enabled = utils.Pointer(true)
			c.LocalStorage.LocalPath = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/ReclaimPolicy/AllowChange", true, "Retain", "Delete", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.ReclaimPolicy = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/ReclaimPolicy/PreventChangeIfEnabled", false, "Retain", "Delete", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.Enabled = utils.Pointer(true)
			c.LocalStorage.LocalPath = utils.Pointer("path")
			c.LocalStorage.ReclaimPolicy = utils.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/Default/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.LocalStorage.Default = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("LocalStorage/Default/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.LocalStorage.Default = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("MetricsServer/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.MetricsServer.Enabled = utils.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("MetricsServer/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.MetricsServer.Enabled = utils.Pointer(v.(bool)) }),
	} {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				result, err := types.MergeClusterConfig(tc.old, tc.new)
				if tc.expectErr {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
					g.Expect(result).To(Equal(tc.expectResult))
				}
			})
		}
	}
}

func TestMergeClusterConfig_Scenarios(t *testing.T) {
	for _, tc := range []struct {
		name         string
		old          types.ClusterConfig
		new          types.ClusterConfig
		expectMerged types.ClusterConfig
		expectErr    bool
	}{
		{
			name: "Kubelet/AllowSetClusterDNS/EnableDNSAfter",
			old: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("1.1.1.1"),
				},
			},
			new: types.ClusterConfig{
				DNS: types.DNS{
					Enabled: utils.Pointer(true),
				},
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("2.2.2.2"),
				},
			},
			expectMerged: types.ClusterConfig{
				DNS: types.DNS{
					Enabled: utils.Pointer(true),
				},
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("2.2.2.2"),
				},
			},
		},
		{
			name: "Kubelet/AllowSetClusterDNS/KeepDNSDisabled",
			old: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("1.1.1.1"),
				},
			},
			new: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("2.2.2.2"),
				},
			},
			expectMerged: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("2.2.2.2"),
				},
			},
		},
		{
			name: "Kubelet/AllowSetClusterDNS/IfDNSEnabledButDNSEmpty",
			old: types.ClusterConfig{
				DNS: types.DNS{
					Enabled: utils.Pointer(true),
				},
			},
			new: types.ClusterConfig{
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("2.2.2.2"),
				},
			},
			expectMerged: types.ClusterConfig{
				DNS: types.DNS{
					Enabled: utils.Pointer(true),
				},
				Kubelet: types.Kubelet{
					ClusterDNS: utils.Pointer("2.2.2.2"),
				},
			},
		},
		{
			name: "LoadBalancer/NeedNetwork",
			old: types.ClusterConfig{
				Network:      types.Network{Enabled: utils.Pointer(true)},
				LoadBalancer: types.LoadBalancer{Enabled: utils.Pointer(true)},
			},
			new: types.ClusterConfig{
				Network: types.Network{Enabled: utils.Pointer(false)},
			},
			expectErr: true,
		},
		{
			name: "LoadBalancer/DisableWithNetwork",
			old: types.ClusterConfig{
				Network:      types.Network{Enabled: utils.Pointer(true)},
				LoadBalancer: types.LoadBalancer{Enabled: utils.Pointer(true)},
			},
			new: types.ClusterConfig{
				Network:      types.Network{Enabled: utils.Pointer(false)},
				LoadBalancer: types.LoadBalancer{Enabled: utils.Pointer(false)},
			},
			expectMerged: types.ClusterConfig{
				Network:      types.Network{Enabled: utils.Pointer(false)},
				LoadBalancer: types.LoadBalancer{Enabled: utils.Pointer(false)},
			},
		},
		{
			name: "LoadBalancer/MissingBGP",
			old: types.ClusterConfig{
				Network:      types.Network{Enabled: utils.Pointer(true)},
				LoadBalancer: types.LoadBalancer{Enabled: utils.Pointer(true)},
			},
			new: types.ClusterConfig{
				LoadBalancer: types.LoadBalancer{BGPMode: utils.Pointer(true)},
			},
			expectErr: true,
		},
		{
			name: "LoadBalancer/InvalidCIDR",
			new: types.ClusterConfig{
				LoadBalancer: types.LoadBalancer{
					CIDRs: utils.Pointer([]string{"not-a-cidr"}),
				},
			},
			expectErr: true,
		},
		{
			name: "LocalStorage/InvalidReclaimPolicy",
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					ReclaimPolicy: utils.Pointer("Invalid"),
				},
			},
			expectErr: true,
		},
		{
			name: "LocalStorage/EnableAndSetPath",
			old: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					LocalPath: utils.Pointer("oldpath"),
				},
			},
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:   utils.Pointer(true),
					LocalPath: utils.Pointer("path"),
				},
			},
			expectMerged: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:   utils.Pointer(true),
					LocalPath: utils.Pointer("path"),
				},
			},
		},
		{
			name: "LocalStorage/EnableAndSetReclaimPolicy",
			old: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					LocalPath:     utils.Pointer("path"),
					ReclaimPolicy: utils.Pointer("Delete"),
				},
			},
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:       utils.Pointer(true),
					ReclaimPolicy: utils.Pointer("Retain"),
				},
			},
			expectMerged: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:       utils.Pointer(true),
					LocalPath:     utils.Pointer("path"),
					ReclaimPolicy: utils.Pointer("Retain"),
				},
			},
		},
		{
			name: "LocalStorage/RequirePath",
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:   utils.Pointer(true),
					LocalPath: utils.Pointer(""),
				},
			},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			// defaults for validation
			tc.old.SetDefaults()
			tc.expectMerged.SetDefaults()

			merged, err := types.MergeClusterConfig(tc.old, tc.new)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(merged).To(Equal(tc.expectMerged))
			}
		})
	}
}
