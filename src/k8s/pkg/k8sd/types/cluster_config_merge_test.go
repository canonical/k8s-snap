package types_test

import (
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
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
		cfg.Network.PodCIDR = vals.Pointer("10.1.0.0/16")
		cfg.Network.ServiceCIDR = vals.Pointer("10.152.183.0/24")
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
		generateMergeClusterConfigTestCases("Certificates/CACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.CACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/CAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.CAKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/FrontProxyCACert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.FrontProxyCACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/FrontProxyCAKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.FrontProxyCAKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/ServiceAccountKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Certificates.ServiceAccountKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/APIServerKubeletClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) {
			c.Certificates.APIServerKubeletClientCert = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Certificates/APIServerKubeletClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) {
			c.Certificates.APIServerKubeletClientKey = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Datastore/Type", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.Type = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqliteCert", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.K8sDqliteCert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqliteKey", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.K8sDqliteKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqlitePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.Datastore.K8sDqlitePort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalServers", true, []string{"localhost:123"}, []string{"localhost:123"}, func(c *types.ClusterConfig, v any) { c.Datastore.ExternalServers = vals.Pointer(v.([]string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalCACert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalCACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalClientCert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalClientKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Network.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Network/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Network.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Network/PodCIDR", false, "10.1.0.0/16", "10.2.0.0/16", func(c *types.ClusterConfig, v any) { c.Network.PodCIDR = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/ServiceCIDR", false, "10.152.183.0/24", "10.152.184.0/24", func(c *types.ClusterConfig, v any) { c.Network.ServiceCIDR = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("APIServer/SecurePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.APIServer.SecurePort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("APIServer/AuthorizationMode", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.AuthorizationMode = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/CloudProvider", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.CloudProvider = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDNS", true, "1.1.1.1", "2.2.2.2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDNS = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDomain", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDomain = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("DNS/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.DNS.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("DNS/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.DNS.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("DNS/UpstreamNameservers", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.DNS.UpstreamNameservers = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Ingress/Enable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = vals.Pointer(true)
			c.Ingress.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Ingress/Disable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = vals.Pointer(true)
			c.Ingress.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Ingress/DefaultTLSSecret", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Ingress.DefaultTLSSecret = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Ingress/EnableProxyProtocol/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Ingress.EnableProxyProtocol = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Ingress/EnableProxyProtocol/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Ingress.EnableProxyProtocol = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Gateway/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = vals.Pointer(true)
			c.Gateway.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Gateway/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = vals.Pointer(true)
			c.Gateway.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = vals.Pointer(true)
			c.LoadBalancer.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Network.Enabled = vals.Pointer(true)
			c.LoadBalancer.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/CIDRs", true, []string{"172.16.101.0/24"}, []string{"172.16.102.0/24"}, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.CIDRs = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/IPRanges", true, []types.LoadBalancer_IPRange{{Start: "10.0.0.10", Stop: "10.0.0.20"}}, []types.LoadBalancer_IPRange{{Start: "10.1.0.10", Stop: "10.1.0.20"}}, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.IPRanges = vals.Pointer(v.([]types.LoadBalancer_IPRange))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/L2Mode/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.LoadBalancer.L2Mode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/L2Mode/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.LoadBalancer.L2Mode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/L2Interfaces", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.L2Interfaces = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPMode/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.BGPMode = vals.Pointer(v.(bool))
			c.LoadBalancer.BGPLocalASN = vals.Pointer(100)
			c.LoadBalancer.BGPPeerAddress = vals.Pointer("10.10.0.0/16")
			c.LoadBalancer.BGPPeerASN = vals.Pointer(101)
			c.LoadBalancer.BGPPeerPort = vals.Pointer(10010)
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPMode/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.BGPMode = vals.Pointer(v.(bool))
			c.LoadBalancer.BGPLocalASN = vals.Pointer(100)
			c.LoadBalancer.BGPPeerAddress = vals.Pointer("10.10.0.0/16")
			c.LoadBalancer.BGPPeerASN = vals.Pointer(101)
			c.LoadBalancer.BGPPeerPort = vals.Pointer(10010)
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPLocalASN", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.LoadBalancer.BGPLocalASN = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPPeerAddress", true, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.LoadBalancer.BGPPeerAddress = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPPeerASN", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.LoadBalancer.BGPPeerASN = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("LoadBalancer/BGPPeerPort", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.LoadBalancer.BGPPeerPort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("LocalStorage/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.LocalStorage.LocalPath = vals.Pointer("path")
			c.LocalStorage.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.LocalStorage.LocalPath = vals.Pointer("path")
			c.LocalStorage.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/LocalPath/AllowChange", true, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.LocalPath = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/LocalPath/PreventChange", false, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.Enabled = vals.Pointer(true)
			c.LocalStorage.LocalPath = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/ReclaimPolicy/AllowChange", true, "Retain", "Delete", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.ReclaimPolicy = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/ReclaimPolicy/PreventChange", false, "Retain", "Delete", func(c *types.ClusterConfig, v any) {
			c.LocalStorage.Enabled = vals.Pointer(true)
			c.LocalStorage.LocalPath = vals.Pointer("path")
			c.LocalStorage.ReclaimPolicy = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("LocalStorage/Default/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.LocalStorage.Default = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("LocalStorage/Default/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.LocalStorage.Default = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("MetricsServer/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.MetricsServer.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("MetricsServer/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.MetricsServer.Enabled = vals.Pointer(v.(bool)) }),
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
			name: "LoadBalancer/NeedNetwork",
			old: types.ClusterConfig{
				Network:      types.Network{Enabled: vals.Pointer(true)},
				LoadBalancer: types.LoadBalancer{Enabled: vals.Pointer(true)},
			},
			new: types.ClusterConfig{
				Network: types.Network{Enabled: vals.Pointer(false)},
			},
			expectErr: true,
		},
		{
			name: "LoadBalancer/DisableWithNetwork",
			old: types.ClusterConfig{
				Network:      types.Network{Enabled: vals.Pointer(true)},
				LoadBalancer: types.LoadBalancer{Enabled: vals.Pointer(true)},
			},
			new: types.ClusterConfig{
				Network:      types.Network{Enabled: vals.Pointer(false)},
				LoadBalancer: types.LoadBalancer{Enabled: vals.Pointer(false)},
			},
			expectMerged: types.ClusterConfig{
				Network:      types.Network{Enabled: vals.Pointer(false)},
				LoadBalancer: types.LoadBalancer{Enabled: vals.Pointer(false)},
			},
		},
		{
			name: "LoadBalancer/MissingBGP",
			old: types.ClusterConfig{
				Network:      types.Network{Enabled: vals.Pointer(true)},
				LoadBalancer: types.LoadBalancer{Enabled: vals.Pointer(true)},
			},
			new: types.ClusterConfig{
				LoadBalancer: types.LoadBalancer{BGPMode: vals.Pointer(true)},
			},
			expectErr: true,
		},
		{
			name: "LoadBalancer/InvalidCIDR",
			new: types.ClusterConfig{
				LoadBalancer: types.LoadBalancer{
					CIDRs: vals.Pointer([]string{"not-a-cidr"}),
				},
			},
			expectErr: true,
		},
		{
			name: "LocalStorage/InvalidReclaimPolicy",
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					ReclaimPolicy: vals.Pointer("Invalid"),
				},
			},
			expectErr: true,
		},
		{
			name: "LocalStorage/EnableAndSetPath",
			old: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					LocalPath: vals.Pointer("oldpath"),
				},
			},
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:   vals.Pointer(true),
					LocalPath: vals.Pointer("path"),
				},
			},
			expectMerged: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:   vals.Pointer(true),
					LocalPath: vals.Pointer("path"),
				},
			},
		},
		{
			name: "LocalStorage/EnableAndSetReclaimPolicy",
			old: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					LocalPath:     vals.Pointer("path"),
					ReclaimPolicy: vals.Pointer("Delete"),
				},
			},
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:       vals.Pointer(true),
					ReclaimPolicy: vals.Pointer("Retain"),
				},
			},
			expectMerged: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:       vals.Pointer(true),
					LocalPath:     vals.Pointer("path"),
					ReclaimPolicy: vals.Pointer("Retain"),
				},
			},
		},
		{
			name: "LocalStorage/RequirePath",
			new: types.ClusterConfig{
				LocalStorage: types.LocalStorage{
					Enabled:   vals.Pointer(true),
					LocalPath: vals.Pointer(""),
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
