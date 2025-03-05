package gateway

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	GatewayDeleteFailedMsgTmpl = "Failed to delete Cilium Gateway, the error was %v"
	GatewayDeployFailedMsgTmpl = "Failed to deploy Cilium Gateway, the error was %v"
)

// ApplyGateway assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the Gateway API CRDs on the cluster and enable the GatewayAPI controllers on Cilium, when gateway.Enabled is true.
// ApplyGateway will remove the Gateway API CRDs from the cluster and disable the GatewayAPI controllers on Cilium, when gateway.Enabled is false.
// ApplyGateway will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyGateway will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyGateway returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func (r reconciler) Reconcile(ctx context.Context, cfg types.ClusterConfig) (types.FeatureStatus, error) {
	gateway := cfg.Gateway
	network := cfg.Network

	if gateway.GetEnabled() {
		return r.enableGateway(ctx, gateway)
	}
	return r.disableGateway(ctx, network)
}

func (r reconciler) enableGateway(ctx context.Context, gateway types.Gateway) (types.FeatureStatus, error) {
	ciliumAgentImageTag := cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag

	helmClient := r.HelmClient()
	snap := r.Snap()

	// Install Gateway API CRDs
	if _, err := helmClient.Apply(ctx, FeatureGateway.GetChart(GatewayChartName), helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := helmClient.Apply(ctx, FeatureGateway.GetChart(GatewayClassChartName), helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	var ciliumValues CiliumValues = map[string]any{}

	if err := ciliumValues.applyClusterConfiguration(gateway); err != nil {
		err = fmt.Errorf("failed to apply cluster configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	changed, err := helmClient.Apply(ctx, cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnly, ciliumValues)
	if err != nil {
		err = fmt.Errorf("failed to upgrade Gateway API cilium configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	if !changed {
		return types.FeatureStatus{
			Enabled: true,
			Version: ciliumAgentImageTag,
			Message: cilium.EnabledMsg,
		}, nil
	}

	if err := cilium.RolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to enable Gateway API: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ciliumAgentImageTag,
		Message: cilium.EnabledMsg,
	}, nil
}

func (r reconciler) disableGateway(ctx context.Context, network types.Network) (types.FeatureStatus, error) {
	ciliumAgentImageTag := cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag

	helmClient := r.HelmClient()
	snap := r.Snap()

	// Delete our GatewayClass named ck-gateway
	if _, err := helmClient.Apply(ctx, FeatureGateway.GetChart(GatewayClassChartName), helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to delete Gateway API GatewayClass: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	var ciliumValues CiliumValues = map[string]any{}

	if err := ciliumValues.applyDisableConfiguration(); err != nil {
		err = fmt.Errorf("failed to apply disable configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	changed, err := helmClient.Apply(ctx, cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), ciliumValues)
	if err != nil {
		err = fmt.Errorf("failed to delete Gateway API cilium configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	// Remove Gateway CRDs if the Gateway feature is disabled.
	// This is done after the Cilium update as cilium requires the CRDs to be present for cleanups.
	if _, err := helmClient.Apply(ctx, FeatureGateway.GetChart(GatewayChartName), helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to delete Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	if !changed {
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: cilium.DisabledMsg,
		}, nil
	}

	if err := cilium.RolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to disable Gateway API: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: false,
		Version: ciliumAgentImageTag,
		Message: cilium.DisabledMsg,
	}, nil
}
