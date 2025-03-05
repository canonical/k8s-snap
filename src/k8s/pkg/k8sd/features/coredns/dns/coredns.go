package dns

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
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
func (r reconciler) Reconcile(ctx context.Context, cfg types.ClusterConfig) (types.FeatureStatus, error) {
	coreDNSImage := FeatureDNS.GetImage(CoreDNSImageName)

	dns := cfg.DNS
	kubelet := cfg.Kubelet

	helmClient := r.HelmClient()

	if !dns.GetEnabled() {
		if _, err := helmClient.Apply(ctx, FeatureDNS.GetChart(CoreDNSChartName), helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall coredns: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: coreDNSImage.Tag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: disabledMsg,
		}, nil
	}

	var values Values = map[string]any{}

	if err := values.ApplyImageOverrides(); err != nil {
		err = fmt.Errorf("failed to apply image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyClusterConfiguration(dns, kubelet); err != nil {
		err = fmt.Errorf("failed to apply cluster configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if _, err := helmClient.Apply(ctx, FeatureDNS.GetChart(CoreDNSChartName), helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to apply coredns: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	dnsIP, err := r.updateClusterDNSandNotify(ctx)
	if err != nil {
		err = fmt.Errorf("failed to update cluster dns: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: coreDNSImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: coreDNSImage.Tag,
		Message: fmt.Sprintf(enabledMsgTmpl, dnsIP),
	}, err
}

func (r reconciler) updateClusterDNSandNotify(ctx context.Context) (string, error) {
	client, err := r.Snap().KubernetesClient("")
	if err != nil {
		return "", fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	dnsIP, err := client.GetServiceClusterIP(ctx, "coredns", "kube-system")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve the coredns service: %w", err)
	}

	if err := UpdateClusterDNS(ctx, r.State(), dnsIP); err != nil {
		return "", fmt.Errorf("failed to update cluster dns: %w", err)
	}

	// DNS IP has changed, notify node config controller
	if err := r.NotifyUpdateNodeConfigController(); err != nil {
		return "", fmt.Errorf("failed to notify update node config controller: %w", err)
	}

	return dnsIP, nil
}

var UpdateClusterDNS = func(ctx context.Context, s state.State, dnsIP string) error {
	if err := s.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := database.SetClusterConfig(ctx, tx, types.ClusterConfig{
			Kubelet: types.Kubelet{ClusterDNS: utils.Pointer(dnsIP)},
		}); err != nil {
			return fmt.Errorf("failed to update cluster configuration for dns=%s: %w", dnsIP, err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("database transaction to update cluster configuration failed: %w", err)
	}

	return nil
}
