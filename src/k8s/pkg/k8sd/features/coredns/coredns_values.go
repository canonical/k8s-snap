package coredns

import (
	"fmt"
	"strings"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type dnsValues map[string]any

func (v *dnsValues) applyDefaults() error {
	values := dnsValues{
		"service": map[string]any{
			"name": "coredns",
		},
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "coredns",
		},
		"deployment": map[string]any{
			"name": "coredns",
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *dnsValues) applyImages() error {
	values := dnsValues{
		"image": map[string]any{
			"repository": imageRepo,
			"tag":        ImageTag,
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}

func (v *dnsValues) applyClusterConfig(dns types.DNS, kubelet types.Kubelet) error {
	values := dnsValues{
		"service": map[string]any{
			"clusterIP": kubelet.GetClusterDNS(),
		},
		"servers": []map[string]any{
			{
				"zones": []map[string]any{
					{"zone": "."},
				},
				"port": 53,
				"plugins": []map[string]any{
					{"name": "errors"},
					{"name": "health", "configBlock": "lameduck 5s"},
					{"name": "ready"},
					{
						"name":        "kubernetes",
						"parameters":  fmt.Sprintf("%s in-addr.arpa ip6.arpa", kubelet.GetClusterDomain()),
						"configBlock": "pods insecure\nfallthrough in-addr.arpa ip6.arpa\nttl 30",
					},
					{"name": "prometheus", "parameters": "0.0.0.0:9153"},
					{"name": "forward", "parameters": fmt.Sprintf(". %s", strings.Join(dns.GetUpstreamNameservers(), " "))},
					{"name": "cache", "parameters": "30"},
					{"name": "loop"},
					{"name": "reload"},
					{"name": "loadbalance"},
				},
			},
		},
	}

	if err := mergo.Merge(v, values, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge: %w", err)
	}

	return nil
}
