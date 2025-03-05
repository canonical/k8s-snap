package dns

import (
	"fmt"
	"strings"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

type Values map[string]any

func (v Values) ApplyImageOverrides(manifest types.FeatureManifest) error {
	coreDNSImage := manifest.GetImage(CoreDNSImageName)

	values := map[string]any{
		"image": map[string]any{
			"repository": coreDNSImage.GetURI(),
			"tag":        coreDNSImage.Tag,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

func (v Values) applyClusterConfiguration(dns types.DNS, kubelet types.Kubelet) error {
	values := map[string]any{
		"service": map[string]any{
			"name":      "coredns",
			"clusterIP": kubelet.GetClusterDNS(),
		},
		"serviceAccount": map[string]any{
			"create": true,
			"name":   "coredns",
		},
		"deployment": map[string]any{
			"name": "coredns",
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

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge cluster configuration: %w", err)
	}

	return nil
}
