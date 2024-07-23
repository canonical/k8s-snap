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
	status := types.FeatureStatus{
		Version: ciliumAgentImageTag,
		Enabled: gateway.GetEnabled(),
	}
	m := snap.HelmClient()

	if _, err := m.Apply(ctx, chartGateway, helm.StatePresentOrDeleted(gateway.GetEnabled()), nil); err != nil {
		if gateway.GetEnabled() {
			enableErr := fmt.Errorf("failed to install Gateway API CRDs: %w", err)
			status.Message = fmt.Sprintf(gatewayDeployFailedMsgTmpl, enableErr)
			return status, enableErr
		} else {
			disableErr := fmt.Errorf("failed to delete Gateway API CRDs: %w", err)
			status.Message = fmt.Sprintf(gatewayDeleteFailedMsgTmpl, disableErr)
			return status, disableErr
		}
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, chartGatewayClass, helm.StatePresentOrDeleted(gateway.GetEnabled()), nil); err != nil {
		if gateway.GetEnabled() {
			enableErr := fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
			status.Message = fmt.Sprintf(gatewayDeployFailedMsgTmpl, enableErr)
			return status, enableErr
		} else {
			disableErr := fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
			status.Message = fmt.Sprintf(gatewayDeleteFailedMsgTmpl, disableErr)
			return status, disableErr
		}
	}

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), map[string]any{"gatewayAPI": map[string]any{"enabled": gateway.GetEnabled()}})
	if err != nil {
		if gateway.GetEnabled() {
			enableErr := fmt.Errorf("failed to apply Gateway API cilium configuration: %w", err)
			status.Message = fmt.Sprintf(gatewayDeployFailedMsgTmpl, enableErr)
			return status, enableErr
		} else {
			disableErr := fmt.Errorf("failed to apply Gateway API cilium configuration: %w", err)
			status.Message = fmt.Sprintf(gatewayDeleteFailedMsgTmpl, disableErr)
			return status, disableErr
		}
	}

	if !changed {
		if gateway.GetEnabled() {
			status.Message = enabledMsg
			return status, nil
		} else {
			status.Message = disabledMsg
			status.Version = ""
			return status, nil
		}
	}

	if !gateway.GetEnabled() {
		status.Message = disabledMsg
		status.Version = ""
		return status, nil
	}
	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		resErr := fmt.Errorf("failed to rollout restart cilium to apply Gateway API: %w", err)
		status.Message = fmt.Sprintf(gatewayDeployFailedMsgTmpl, resErr)
		return status, resErr
	}

	status.Message = enabledMsg
	return status, nil
}
