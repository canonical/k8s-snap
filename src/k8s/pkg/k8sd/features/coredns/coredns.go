package coredns

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	enabledMsgTmpl      = "enabled at %s"
	disabledMsg         = "disabled"
	deleteFailedMsgTmpl = "Failed to delete DNS, the error was: %v"
	deployFailedMsgTmpl = "Failed to deploy DNS, the error was: %v"
)

// ApplyDNS manages the deployment of CoreDNS, with customization options from dns and kubelet, which are retrieved from the cluster configuration.
// ApplyDNS will uninstall CoreDNS from the cluster if dns.Enabled is false.
// ApplyDNS will install or refresh CoreDNS if dns.Enabled is true.
// ApplyDNS will return the ClusterIP address of the coredns service, if successful.
// ApplyDNS will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyDNS returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyDNS(ctx context.Context, snap snap.Snap, dns types.DNS, kubelet types.Kubelet, _ types.Annotations) (types.FeatureStatus, string, error) {
	m := snap.HelmClient()

	if !dns.GetEnabled() {
		if _, err := m.Apply(ctx, Chart, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall coredns: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ImageTag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, "", err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: disabledMsg,
		}, "", nil
	}

	values := map[string]any{
		"image": map[string]any{
			"repository": imageRepo,
			"tag":        ImageTag,
		},
		"service": map[string]any{
			"name":      "coredns",
			"clusterIP": kubelet.GetClusterDNS(),
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

	if _, err := m.Apply(ctx, Chart, helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to apply coredns: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		err = fmt.Errorf("failed to create kubernetes client: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}
	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		err = fmt.Errorf("failed to retrieve the coredns service: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ImageTag,
		Message: fmt.Sprintf(enabledMsgTmpl, dnsIP),
	}, dnsIP, err
}
