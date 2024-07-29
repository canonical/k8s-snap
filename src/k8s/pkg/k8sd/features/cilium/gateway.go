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
	m := snap.HelmClient()

	if _, err := m.Apply(ctx, chartGateway, helm.StatePresentOrDeleted(gateway.GetEnabled()), nil); err != nil {
		if gateway.GetEnabled() {
			err = fmt.Errorf("failed to install Gateway API CRDs: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to delete Gateway API CRDs: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(gatewayDeleteFailedMsgTmpl, err),
			}, err
		}
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, chartGatewayClass, helm.StatePresentOrDeleted(gateway.GetEnabled()), nil); err != nil {
		if gateway.GetEnabled() {
			err = fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(gatewayDeleteFailedMsgTmpl, err),
			}, err
		}
	}

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), map[string]any{"gatewayAPI": map[string]any{"enabled": gateway.GetEnabled()}})
	if err != nil {
		if gateway.GetEnabled() {
			err = fmt.Errorf("failed to apply Gateway API cilium configuration: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(gatewayDeployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to apply Gateway API cilium configuration: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(gatewayDeleteFailedMsgTmpl, err),
			}, err
		}
	}

	if !changed {
		if gateway.GetEnabled() {
			return types.FeatureStatus{
				Enabled: true,
				Version: ciliumAgentImageTag,
				Message: enabledMsg,
			}, nil
		} else {
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: disabledMsg,
			}, nil
		}
	}

	if !gateway.GetEnabled() {
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: disabledMsg,
		}, nil
	}
	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to apply Gateway API: %w", err)
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
