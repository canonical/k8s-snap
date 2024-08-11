package k8s

import (
	"fmt"
	"testing"

	apiv1 "github.com/canonical/k8s-snap-api-v1/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

type mapstructureTestCase struct {
	name       string
	val        string
	expectErr  bool
	assertions []types.GomegaMatcher
}

func generateMapstructureTestCasesBool(keyName string, fieldName string) []mapstructureTestCase {
	return []mapstructureTestCase{
		{
			val:        fmt.Sprintf("%s=true", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer(true))},
		},
		{
			val:        fmt.Sprintf("%s=false", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer(false))},
		},
		{
			val:        fmt.Sprintf("%s=", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer(false))},
		},
		{
			val:       fmt.Sprintf("%s=yes", keyName),
			expectErr: true,
		},
	}
}

func generateMapstructureTestCasesStringSlice(keyName string, fieldName string) []mapstructureTestCase {
	return []mapstructureTestCase{
		{
			val:        fmt.Sprintf("%s=", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{}))},
		},
		{
			val:        fmt.Sprintf("%s=[]", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{}))},
		},
		{
			val:        fmt.Sprintf("%s=100", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"100"}))},
		},
		{
			val:        fmt.Sprintf("%s=t1", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"t1"}))},
		},
		{
			val:        fmt.Sprintf(`%s=["t1"]`, keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"t1"}))},
		},
		{
			val:        fmt.Sprintf("%s=[t1]", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"t1"}))},
		},
		{
			val:        fmt.Sprintf("%s=t1, t2", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"t1", "t2"}))},
		},
		{
			val:        fmt.Sprintf(`%s=["t1", "t2"]`, keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"t1", "t2"}))},
		},
		{
			val:        fmt.Sprintf(`%s=[t1, t2]`, keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer([]string{"t1", "t2"}))},
		},
	}
}

func generateMapstructureTestCasesMap(keyName string, fieldName string) []mapstructureTestCase {
	return []mapstructureTestCase{
		{
			val:        fmt.Sprintf("%s=", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{})},
		},
		{
			val:        fmt.Sprintf("%s={}", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{})},
		},
		{
			val:        fmt.Sprintf("%s=k1=", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{"k1": ""})},
		},
		{
			val:        fmt.Sprintf("%s=k1=,k2=test", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{"k1": "", "k2": "test"})},
		},
		{
			val:        fmt.Sprintf("%s=k1=v1", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{"k1": "v1"})},
		},
		{
			val:        fmt.Sprintf("%s=k1=v1,k2=v2", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{"k1": "v1", "k2": "v2"})},
		},
		{
			val:        fmt.Sprintf("%s={k1: v1}", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{"k1": "v1"})},
		},
		{
			val:        fmt.Sprintf("%s={k1: v1, k2: v2}", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, map[string]string{"k1": "v1", "k2": "v2"})},
		},
		{
			val:       fmt.Sprintf("%s=k1,k2", keyName),
			expectErr: true,
		},
	}
}

func generateMapstructureTestCasesString(keyName string, fieldName string) []mapstructureTestCase {
	return []mapstructureTestCase{
		{
			val:        fmt.Sprintf("%s=", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer(""))},
		},
		{
			val:        fmt.Sprintf("%s=t1", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer("t1"))},
		},
	}
}

func generateMapstructureTestCasesInt(keyName string, fieldName string) []mapstructureTestCase {
	return []mapstructureTestCase{
		{
			val:        fmt.Sprintf("%s=", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer(0))},
		},
		{
			val:        fmt.Sprintf("%s=100", keyName),
			assertions: []types.GomegaMatcher{HaveField(fieldName, utils.Pointer(100))},
		},
		{
			val:       fmt.Sprintf("%s=notanumber", keyName),
			expectErr: true,
		},
	}
}

func Test_updateConfigMapstructure(t *testing.T) {
	for _, tcs := range [][]mapstructureTestCase{
		generateMapstructureTestCasesBool("dns.enabled", "DNS.Enabled"),
		generateMapstructureTestCasesBool("gateway.enabled", "Gateway.Enabled"),
		generateMapstructureTestCasesBool("ingress.enable-proxy-protocol", "Ingress.EnableProxyProtocol"),
		generateMapstructureTestCasesBool("ingress.enabled", "Ingress.Enabled"),
		generateMapstructureTestCasesBool("load-balancer.bgp-mode", "LoadBalancer.BGPMode"),
		generateMapstructureTestCasesBool("load-balancer.l2-mode", "LoadBalancer.L2Mode"),
		generateMapstructureTestCasesBool("load-balancer.enabled", "LoadBalancer.Enabled"),
		generateMapstructureTestCasesBool("load-balancer.enabled", "LoadBalancer.Enabled"),
		generateMapstructureTestCasesBool("local-storage.default", "LocalStorage.Default"),
		generateMapstructureTestCasesBool("local-storage.enabled", "LocalStorage.Enabled"),
		generateMapstructureTestCasesBool("metrics-server.enabled", "MetricsServer.Enabled"),
		generateMapstructureTestCasesBool("network.enabled", "Network.Enabled"),

		generateMapstructureTestCasesString("cloud-provider", "CloudProvider"),
		generateMapstructureTestCasesString("dns.cluster-domain", "DNS.ClusterDomain"),
		generateMapstructureTestCasesString("dns.service-ip", "DNS.ServiceIP"),
		generateMapstructureTestCasesString("ingress.default-tls-secret", "Ingress.DefaultTLSSecret"),
		generateMapstructureTestCasesString("load-balancer.bgp-peer-address", "LoadBalancer.BGPPeerAddress"),
		generateMapstructureTestCasesString("local-storage.local-path", "LocalStorage.LocalPath"),
		generateMapstructureTestCasesString("local-storage.reclaim-policy", "LocalStorage.ReclaimPolicy"),

		generateMapstructureTestCasesStringSlice("dns.upstream-nameservers", "DNS.UpstreamNameservers"),
		generateMapstructureTestCasesStringSlice("load-balancer.cidrs", "LoadBalancer.CIDRs"),
		generateMapstructureTestCasesStringSlice("load-balancer.l2-interfaces", "LoadBalancer.L2Interfaces"),

		generateMapstructureTestCasesInt("load-balancer.bgp-local-asn", "LoadBalancer.BGPLocalASN"),
		generateMapstructureTestCasesInt("load-balancer.bgp-peer-asn", "LoadBalancer.BGPPeerASN"),
		generateMapstructureTestCasesInt("load-balancer.bgp-peer-port", "LoadBalancer.BGPPeerPort"),

		generateMapstructureTestCasesMap("annotations", "Annotations"),
	} {
		for _, tc := range tcs {
			t.Run(tc.val, func(t *testing.T) {
				g := NewWithT(t)

				var cfg apiv1.UserFacingClusterConfig
				err := updateConfigMapstructure(&cfg, tc.val)
				if tc.expectErr {
					g.Expect(err).To(HaveOccurred())
				} else {
					g.Expect(err).To(BeNil())
					g.Expect(cfg).To(SatisfyAll(tc.assertions...))
				}
			})
		}
	}
}
