package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/state"
)

const (
	GatewayDeleteFailedMsgTmpl = "Failed to delete Cilium Gateway, the error was %v"
	GatewayDeployFailedMsgTmpl = "Failed to deploy Cilium Gateway, the error was %v"
)

const GATEWAY_VERSION = "v1.0.0"

// ApplyGateway assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyGateway will deploy the Gateway API CRDs on the cluster and enable the GatewayAPI controllers on Cilium, when gateway.Enabled is true.
// ApplyGateway will remove the Gateway API CRDs from the cluster and disable the GatewayAPI controllers on Cilium, when gateway.Enabled is false.
// ApplyGateway will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyGateway will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyGateway returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyGateway(ctx context.Context, _ state.State, snap snap.Snap, gateway types.Gateway, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	if gateway.GetEnabled() {
		return enableGateway(ctx, snap)
	}
	return disableGateway(ctx, snap, network)
}

func enableGateway(ctx context.Context, snap snap.Snap) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	// Install Gateway API CRDs
	if _, err := m.Apply(ctx, features.Gateway, GATEWAY_VERSION, chartGateway, helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, features.Gateway, GATEWAY_VERSION, chartGatewayClass, helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	values := gatewayValues{}

	if err := values.applyDefaults(); err != nil {
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	parent := helm.FeatureMeta{
		FeatureName: features.Network,
		Version:     NETWORK_VERSION,
		Chart:       ChartCilium,
	}

	sub := helm.PseudoFeatureMeta{
		FeatureName: features.Gateway,
		Version:     GATEWAY_VERSION,
	}

	changed, err := m.ApplyDependent(ctx, parent, sub, helm.StatePresent, values)
	if err != nil {
		err = fmt.Errorf("failed to upgrade Gateway API cilium configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	if !changed {
		return types.FeatureStatus{
			Enabled: true,
			Version: CiliumAgentImageTag,
			Message: EnabledMsg,
		}, nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to enable Gateway API: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: CiliumAgentImageTag,
		Message: EnabledMsg,
	}, nil
}

func disableGateway(ctx context.Context, snap snap.Snap, network types.Network) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	// Delete our GatewayClass named ck-gateway
	if _, err := m.Apply(ctx, features.Gateway, GATEWAY_VERSION, chartGatewayClass, helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to delete Gateway API GatewayClass: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	values := gatewayValues{}

	if err := values.applyDisable(); err != nil {
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	parent := helm.FeatureMeta{
		FeatureName: features.Network,
		Version:     NETWORK_VERSION,
		Chart:       ChartCilium,
	}

	sub := helm.PseudoFeatureMeta{
		FeatureName: features.Gateway,
		Version:     GATEWAY_VERSION,
	}

	changed, err := m.ApplyDependent(ctx, parent, sub, helm.StateDeleted, values)
	if err != nil {
		err = fmt.Errorf("failed to delete Gateway API cilium configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	// Remove Gateway CRDs if the Gateway feature is disabled.
	// This is done after the Cilium update as cilium requires the CRDs to be present for cleanups.
	if _, err := m.Apply(ctx, features.Gateway, GATEWAY_VERSION, chartGateway, helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to delete Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
		}, err
	}

	if !changed {
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: DisabledMsg,
		}, nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to disable Gateway API: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: false,
		Version: CiliumAgentImageTag,
		Message: DisabledMsg,
	}, nil
}
