package newtypes_test

import (
	"fmt"
	"testing"

	newtypes "github.com/canonical/k8s/pkg/k8sd/new"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

type mergeClusterConfigTestCase struct {
	name         string
	old          newtypes.ClusterConfig
	new          newtypes.ClusterConfig
	expectResult newtypes.ClusterConfig
	expectErr    bool
}

func generateMergeClusterConfigTestCases[T any](field string, changeAllowed bool, val1 T, val2 T, update func(*newtypes.ClusterConfig, any)) []mergeClusterConfigTestCase {
	var cfgNil, cfgZero, cfgOne, cfgTwo newtypes.ClusterConfig
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
		generateMergeClusterConfigTestCases("Certificates/CACert", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Certificates.CACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/CAKey", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Certificates.CAKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/FrontProxyCACert", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Certificates.FrontProxyCACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/FrontProxyCAKey", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Certificates.FrontProxyCAKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/ServiceAccountKey", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Certificates.ServiceAccountKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Certificates/APIServerKubeletClientCert", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) {
			c.Certificates.APIServerKubeletClientCert = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Certificates/APIServerKubeletClientKey", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) {
			c.Certificates.APIServerKubeletClientKey = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Datastore/Type", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.Type = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqliteCert", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.K8sDqliteCert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqliteKey", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.K8sDqliteKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/K8sDqlitePort", false, 6443, 16443, func(c *newtypes.ClusterConfig, v any) { c.Datastore.K8sDqlitePort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalURL", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.ExternalURL = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalCACert", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.ExternalCACert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientCert", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.ExternalClientCert = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Datastore/ExternalClientKey", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Datastore.ExternalClientKey = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/PodCIDR", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Network.PodCIDR = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Network/ServiceCIDR", false, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Network.ServiceCIDR = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("APIServer/SecurePort", false, 6443, 16443, func(c *newtypes.ClusterConfig, v any) { c.APIServer.SecurePort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("APIServer/AuthorizationMode", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.APIServer.AuthorizationMode = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Features/Network/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.Network.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/Network/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.Network.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/DNS/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.DNS.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/DNS/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.DNS.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/DNS/UpstreamNameservers", true, []string{"c1"}, []string{"c2"}, func(c *newtypes.ClusterConfig, v any) {
			c.Features.DNS.UpstreamNameservers = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Features/Ingress/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.Ingress.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/Ingress/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.Ingress.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/Ingress/DefaultTLSSecret", true, "v1", "v2", func(c *newtypes.ClusterConfig, v any) { c.Features.Ingress.DefaultTLSSecret = vals.Pointer(v.(string)) }),
		generateMergeClusterConfigTestCases("Features/Ingress/EnableProxyProtocol/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) {
			c.Features.Ingress.EnableProxyProtocol = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Ingress/EnableProxyProtocol/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) {
			c.Features.Ingress.EnableProxyProtocol = vals.Pointer(v.(bool))
		}),
		generateMergeClusterConfigTestCases("Features/Gateway/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.Gateway.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/Gateway/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.Gateway.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/CIDRs", true, []string{"c1"}, []string{"c2"}, func(c *newtypes.ClusterConfig, v any) {
			c.Features.LoadBalancer.CIDRs = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/L2Mode/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.L2Mode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/L2Mode/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.L2Mode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/L2Interfaces", true, []string{"c1"}, []string{"c2"}, func(c *newtypes.ClusterConfig, v any) {
			c.Features.LoadBalancer.L2Interfaces = vals.Pointer(v.([]string))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPMode/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.BGPMode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPMode/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.BGPMode = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPLocalASN", true, 6443, 16443, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.BGPLocalASN = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPPeerAddress", true, "a1", "a2", func(c *newtypes.ClusterConfig, v any) {
			c.Features.LoadBalancer.BGPPeerAddress = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPPeerASN", true, 6443, 16443, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.BGPPeerASN = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Features/LoadBalancer/BGPPeerPort", true, 6443, 16443, func(c *newtypes.ClusterConfig, v any) { c.Features.LoadBalancer.BGPPeerPort = vals.Pointer(v.(int)) }),
		generateMergeClusterConfigTestCases("Features/LocalStorage/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.LocalStorage.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LocalStorage/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.LocalStorage.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LocalStorage/LocalPath", false, "a1", "a2", func(c *newtypes.ClusterConfig, v any) {
			c.Features.LocalStorage.LocalPath = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/ReclaimPolicy", true, "a1", "a2", func(c *newtypes.ClusterConfig, v any) {
			c.Features.LocalStorage.ReclaimPolicy = vals.Pointer(v.(string))
		}),
		generateMergeClusterConfigTestCases("Features/LocalStorage/SetDefault/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.LocalStorage.SetDefault = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/LocalStorage/SetDefault/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.LocalStorage.SetDefault = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/MetricsServer/Enable", true, true, false, func(c *newtypes.ClusterConfig, v any) { c.Features.MetricsServer.Enabled = vals.Pointer(v.(bool)) }),
		generateMergeClusterConfigTestCases("Features/MetricsServer/Disable", true, false, true, func(c *newtypes.ClusterConfig, v any) { c.Features.MetricsServer.Enabled = vals.Pointer(v.(bool)) }),
	} {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)

				result, err := newtypes.MergeClusterConfig(tc.old, tc.new)
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
