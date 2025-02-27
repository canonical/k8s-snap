package dns

import (
	"context"
	"fmt"

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
func ApplyDNS(ctx context.Context, snap snap.Snap, m helm.Client, dns types.DNS, kubelet types.Kubelet, _ types.Annotations) (types.FeatureStatus, string, error) {
	coreDNSImage := FeatureDNS.GetImage(CoreDNSImageName)

	if !dns.GetEnabled() {
		if _, err := m.Apply(ctx, FeatureDNS.GetChart(CoreDNSChartName), helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall coredns: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: coreDNSImage.Tag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, "", err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: disabledMsg,
		}, "", nil
	}

	var values Values = map[string]any{}

	if err := values.ApplyImageOverrides(); err != nil {
		err = fmt.Errorf("failed to apply image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}

	if err := values.applyClusterConfiguration(dns, kubelet); err != nil {
		err = fmt.Errorf("failed to apply cluster configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}

	if _, err := m.Apply(ctx, FeatureDNS.GetChart(CoreDNSChartName), helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to apply coredns: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		err = fmt.Errorf("failed to create kubernetes client: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}
	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		err = fmt.Errorf("failed to retrieve the coredns service: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, "", err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: coreDNSImage.Tag,
		Message: fmt.Sprintf(enabledMsgTmpl, dnsIP),
	}, dnsIP, err
}
