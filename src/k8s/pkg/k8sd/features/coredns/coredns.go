package coredns

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns/internal"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

const (
	enabledMsgTmpl      = "enabled at %s"
	disabledMsg         = "disabled"
	deleteFailedMsgTmpl = "Failed to delete DNS, the error was: %v"
	deployFailedMsgTmpl = "Failed to deploy DNS, the error was: %v"
)

const DNS_VERSION = "v1.0.0"

// ApplyDNS manages the deployment of CoreDNS, with customization options from dns and kubelet, which are retrieved from the cluster configuration.
// ApplyDNS will uninstall CoreDNS from the cluster if dns.Enabled is false.
// ApplyDNS will install or refresh CoreDNS if dns.Enabled is true.
// ApplyDNS will return the ClusterIP address of the coredns service, if successful.
// ApplyDNS will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyDNS returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyDNS(ctx context.Context, s state.State, snap snap.Snap, dns types.DNS, kubelet types.Kubelet, _ types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	if !dns.GetEnabled() {
		if _, err := m.Apply(ctx, features.DNS, DNS_VERSION, Chart, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall coredns: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ImageTag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
		if err := internal.UpdateClusterDNS(ctx, s, ""); err != nil {
			err = fmt.Errorf("failed to update cluster DNS: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ImageTag,
				Message: fmt.Sprintf(deployFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: disabledMsg,
		}, nil
	}

	values := dnsValues{}

	if err := values.applyDefaults(); err != nil {
		err = fmt.Errorf("failed to apply defaults: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyImages(); err != nil {
		err = fmt.Errorf("failed to apply images: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyClusterConfig(dns, kubelet); err != nil {
		err = fmt.Errorf("failed to apply cluster config: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if _, err := m.Apply(ctx, features.DNS, DNS_VERSION, Chart, helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to apply coredns: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		err = fmt.Errorf("failed to create kubernetes client: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		err = fmt.Errorf("failed to retrieve the coredns service: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := internal.UpdateClusterDNS(ctx, s, dnsIP); err != nil {
		err = fmt.Errorf("failed to update cluster DNS: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ImageTag,
		Message: fmt.Sprintf(enabledMsgTmpl, dnsIP),
	}, err
}
