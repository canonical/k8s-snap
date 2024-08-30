package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	gatewayDeleteFailedMsgTmpl = "Failed to delete Cilium Gateway, the error was %v"
	gatewayDeployFailedMsgTmpl = "Failed to deploy Cilium Gateway, the error was %v"
)

// ApplyGateway assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the Gateway API CRDs on the cluster and enable the GatewayAPI controllers on Cilium, when gateway.Enabled is true.
// ApplyGateway will remove the Gateway API CRDs from the cluster and disable the GatewayAPI controllers on Cilium, when gateway.Enabled is false.
// ApplyGateway will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyGateway will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyGateway returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyGateway(ctx context.Context, snap snap.Snap, gateway types.Gateway, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	if gateway.GetEnabled() {
		return enableGateway(ctx, snap)
	}
	return disableGateway(ctx, snap, network)
}

func enableGateway(ctx context.Context, snap snap.Snap) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	// Install Gateway API CRDs
	if _, err := m.Apply(ctx, chartGateway, helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
		}, err
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, chartGatewayClass, helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
		}, err
	}

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnly, map[string]any{"gatewayAPI": map[string]any{"enabled": true}})
	if err != nil {
		err = fmt.Errorf("failed to upgrade Gateway API cilium configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
		}, err
	}

	if !changed {
		return types.FeatureStatus{
			Enabled: true,
			Version: ciliumAgentImageTag,
			Message: enabledMsg,
		}, nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to enable Gateway API: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ciliumAgentImageTag,
		Message: enabledMsg,
	}, nil
}

func disableGateway(ctx context.Context, snap snap.Snap, network types.Network) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	// Delete our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, chartGatewayClass, helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to delete Gateway API GatewayClass: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), map[string]any{"gatewayAPI": map[string]any{"enabled": false}})
	if err != nil {
		err = fmt.Errorf("failed to delete Gateway API cilium configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	// Remove Gateway CRDs if the Gateway feature is disabled.
	// This is done after the Cilium update as cilium requires the CRDs to be present for cleanups.
	if _, err := m.Apply(ctx, chartGateway, helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to delete Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	if !changed {
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: disabledMsg,
		}, nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to disable Gateway API: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: false,
		Version: ciliumAgentImageTag,
		Message: disabledMsg,
	}, nil
}
