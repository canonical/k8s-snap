package features

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/k8s"
)

// ApplyDNS is used to configure the DNS feature on Canonical Kubernetes.
// ApplyDNS manages the deployment of CoreDNS, with customization options from dnsConfig and kubeletConfig which are retrieved from the cluster configuration.
// ApplyDNS will uninstall CoreDNS from the cluster if dnsConfig.Enabled is false.
// ApplyDNS will install or refresh CoreDNS if dnsConfig.Enabled is true.
// ApplyDNS will return the ClusterIP address of the coredns service, if successful.
// ApplyDNS returns an error if anything fails.
func ApplyDNS(ctx context.Context, snap snap.Snap, dnsConfig types.DNS, kubeletConfig types.Kubelet) (string, error) {
	m := newHelm(snap)

	if !dnsConfig.GetEnabled() {
		if _, err := m.Apply(ctx, featureCoreDNS, stateDeleted, nil); err != nil {
			return "", fmt.Errorf("failed to uninstall coredns: %w", err)
		}
		return "", nil
	}

	values := map[string]any{
		"image": map[string]any{
			"repository": dnsImageRepository,
			"tag":        dnsImageTag,
		},
		"service": map[string]any{
			"name":      "coredns",
			"clusterIP": kubeletConfig.GetClusterDNS(),
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
						"parameters":  fmt.Sprintf("%s in-addr.arpa ip6.arpa", kubeletConfig.GetClusterDomain()),
						"configBlock": "pods insecure\nfallthrough in-addr.arpa ip6.arpa\nttl 30",
					},
					{"name": "prometheus", "parameters": "0.0.0.0:9153"},
					{"name": "forward", "parameters": fmt.Sprintf(". %s", strings.Join(dnsConfig.GetUpstreamNameservers(), " "))},
					{"name": "cache", "parameters": "30"},
					{"name": "loop"},
					{"name": "reload"},
					{"name": "loadbalance"},
				},
			},
		},
	}

	if _, err := m.Apply(ctx, featureCoreDNS, statePresent, values); err != nil {
		return "", fmt.Errorf("failed to apply coredns: %w", err)
	}

	client, err := k8s.NewClient(snap.KubernetesRESTClientGetter(""))
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve the coredns service: %w", err)
	}

	return dnsIP, nil
}
