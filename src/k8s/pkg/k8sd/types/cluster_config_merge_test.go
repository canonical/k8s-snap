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
		generateMergeClusterConfigTestCases("Datastore/ExternalURL", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalURL = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalCACert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalCACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientCert", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalClientCert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientKey", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Datastore.ExternalClientKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/PodCIDR", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Network.PodCIDR = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/ServiceCIDR", false, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Network.ServiceCIDR = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("APIServer/SecurePort", false, 6443, 16443, func(c *types.ClusterConfig, v any) { c.APIServer.SecurePort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("APIServer/AuthorizationMode", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.APIServer.AuthorizationMode = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/CloudProvider", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.CloudProvider = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDNS", true, "1.1.1.1", "2.2.2.2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDNS = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Kubelet/ClusterDomain", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Kubelet.ClusterDomain = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Features/Network/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Features.Network.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/Network/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Features.Network.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/DNS/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Features.DNS.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/DNS/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Features.DNS.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/DNS/UpstreamNameservers", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.Features.DNS.UpstreamNameservers = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Features/Ingress/Enable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Features.Network.Enabled = vals.Pointer(true)
			c.Features.Ingress.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Ingress/Disable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Features.Network.Enabled = vals.Pointer(true)
			c.Features.Ingress.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Ingress/DefaultTLSSecret", true, "v1", "v2", func(c *types.ClusterConfig, v any) { c.Features.Ingress.DefaultTLSSecret = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Features/Ingress/EnableProxyProtocol/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Features.Ingress.EnableProxyProtocol = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Ingress/EnableProxyProtocol/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Features.Ingress.EnableProxyProtocol = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Gateway/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Features.Network.Enabled = vals.Pointer(true)
			c.Features.Gateway.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Gateway/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Features.Network.Enabled = vals.Pointer(true)
			c.Features.Gateway.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Features.Network.Enabled = vals.Pointer(true)
			c.Features.LoadBalancer.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Features.Network.Enabled = vals.Pointer(true)
			c.Features.LoadBalancer.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/CIDRs", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.Features.LoadBalancer.CIDRs = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/L2Mode/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Features.LoadBalancer.L2Mode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/L2Mode/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Features.LoadBalancer.L2Mode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/L2Interfaces", true, []string{"c1"}, []string{"c2"}, func(c *types.ClusterConfig, v any) {
			c.Features.LoadBalancer.L2Interfaces = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPMode/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Features.LoadBalancer.BGPMode = vals.Pointer(v.(bool))
			c.Features.LoadBalancer.BGPLocalASN = vals.Pointer(100)
			c.Features.LoadBalancer.BGPPeerAddress = vals.Pointer("10.10.0.0/16")
			c.Features.LoadBalancer.BGPPeerASN = vals.Pointer(101)
			c.Features.LoadBalancer.BGPPeerPort = vals.Pointer(10010)
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPMode/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Features.LoadBalancer.BGPMode = vals.Pointer(v.(bool))
			c.Features.LoadBalancer.BGPLocalASN = vals.Pointer(100)
			c.Features.LoadBalancer.BGPPeerAddress = vals.Pointer("10.10.0.0/16")
			c.Features.LoadBalancer.BGPPeerASN = vals.Pointer(101)
			c.Features.LoadBalancer.BGPPeerPort = vals.Pointer(10010)
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPLocalASN", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.Features.LoadBalancer.BGPLocalASN = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPPeerAddress", true, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.Features.LoadBalancer.BGPPeerAddress = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPPeerASN", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.Features.LoadBalancer.BGPPeerASN = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPPeerPort", true, 6443, 16443, func(c *types.ClusterConfig, v any) { c.Features.LoadBalancer.BGPPeerPort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Features/LocalStorage/Enable", true, true, false, func(c *types.ClusterConfig, v any) {
			c.Features.LocalStorage.LocalPath = vals.Pointer("path")
			c.Features.LocalStorage.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/Disable", true, false, true, func(c *types.ClusterConfig, v any) {
			c.Features.LocalStorage.LocalPath = vals.Pointer("path")
			c.Features.LocalStorage.Enabled = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/LocalPath/AllowChange", true, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.Features.LocalStorage.LocalPath = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/LocalPath/PreventChange", false, "a1", "a2", func(c *types.ClusterConfig, v any) {
			c.Features.LocalStorage.Enabled = vals.Pointer(true)
			c.Features.LocalStorage.LocalPath = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/ReclaimPolicy/AllowChange", true, "Retain", "Delete", func(c *types.ClusterConfig, v any) {
			c.Features.LocalStorage.ReclaimPolicy = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/ReclaimPolicy/PreventChange", false, "Retain", "Delete", func(c *types.ClusterConfig, v any) {
			c.Features.LocalStorage.Enabled = vals.Pointer(true)
			c.Features.LocalStorage.LocalPath = vals.Pointer("path")
			c.Features.LocalStorage.ReclaimPolicy = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/SetDefault/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Features.LocalStorage.SetDefault = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LocalStorage/SetDefault/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Features.LocalStorage.SetDefault = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/MetricsServer/Enable", true, true, false, func(c *types.ClusterConfig, v any) { c.Features.MetricsServer.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/MetricsServer/Disable", true, false, true, func(c *types.ClusterConfig, v any) { c.Features.MetricsServer.Enabled = vals.Pointer(v.(bool)) }),
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
				Features: types.Features{
					Network:      types.NetworkFeature{Enabled: vals.Pointer(true)},
					LoadBalancer: types.LoadBalancerFeature{Enabled: vals.Pointer(true)},
				},
			},
			new: types.ClusterConfig{
				Features: types.Features{
					Network: types.NetworkFeature{Enabled: vals.Pointer(false)},
				},
			},
			expectErr: true,
		},
		{
			name: "LoadBalancer/DisableWithNetwork",
			old: types.ClusterConfig{
				Features: types.Features{
					Network:      types.NetworkFeature{Enabled: vals.Pointer(true)},
					LoadBalancer: types.LoadBalancerFeature{Enabled: vals.Pointer(true)},
				},
			},
			new: types.ClusterConfig{
				Features: types.Features{
					Network:      types.NetworkFeature{Enabled: vals.Pointer(false)},
					LoadBalancer: types.LoadBalancerFeature{Enabled: vals.Pointer(false)},
				},
			},
			expectMerged: types.ClusterConfig{
				Features: types.Features{
					Network:      types.NetworkFeature{Enabled: vals.Pointer(false)},
					LoadBalancer: types.LoadBalancerFeature{Enabled: vals.Pointer(false)},
				},
			},
		},
		{
			name: "LoadBalancer/MissingBGP",
			old: types.ClusterConfig{
				Features: types.Features{
					Network:      types.NetworkFeature{Enabled: vals.Pointer(true)},
					LoadBalancer: types.LoadBalancerFeature{Enabled: vals.Pointer(true)},
				},
			},
			new: types.ClusterConfig{
				Features: types.Features{
					LoadBalancer: types.LoadBalancerFeature{BGPMode: vals.Pointer(true)},
				},
			},
			expectErr: true,
		},
		{
			name: "LocalStorage/InvalidReclaimPolicy",
			new: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						ReclaimPolicy: vals.Pointer("Invalid"),
					},
				},
			},
			expectErr: true,
		},
		{
			name: "LocalStorage/EnableAndSetPath",
			old: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						LocalPath: vals.Pointer("oldpath"),
					},
				},
			},
			new: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						Enabled:   vals.Pointer(true),
						LocalPath: vals.Pointer("path"),
					},
				},
			},
			expectMerged: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						Enabled:   vals.Pointer(true),
						LocalPath: vals.Pointer("path"),
					},
				},
			},
		},
		{
			name: "LocalStorage/EnableAndSetReclaimPolicy",
			old: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						LocalPath:     vals.Pointer("path"),
						ReclaimPolicy: vals.Pointer("Delete"),
					},
				},
			},
			new: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						Enabled:       vals.Pointer(true),
						ReclaimPolicy: vals.Pointer("Retain"),
					},
				},
			},
			expectMerged: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						Enabled:       vals.Pointer(true),
						LocalPath:     vals.Pointer("path"),
						ReclaimPolicy: vals.Pointer("Retain"),
					},
				},
			},
		},
		{
			name: "LocalStorage/RequirePath",
			new: types.ClusterConfig{
				Features: types.Features{
					LocalStorage: types.LocalStorageFeature{
						Enabled: vals.Pointer(true),
					},
				},
			},
			expectErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
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
