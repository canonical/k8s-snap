package gateway

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/microcluster/v2/state"
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
	networkManifest, err := r.getNetworkManifest(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get network manifest: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: "",
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	gateway := cfg.Gateway
	network := cfg.Network

	if gateway.GetEnabled() {
		return r.enableGateway(ctx, networkManifest, gateway)
	}
	return r.disableGateway(ctx, networkManifest, network)
}

func (r reconciler) enableGateway(ctx context.Context, networkManifest *types.FeatureManifest, gateway types.Gateway) (types.FeatureStatus, error) {
	ciliumAgentImageTag := networkManifest.GetImage(cilium_network.CiliumAgentImageName).Tag

	helmClient := r.HelmClient()
	snap := r.Snap()

	// Install Gateway API CRDs
	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(GatewayChartName), helm.StatePresent, nil); err != nil {
		err = fmt.Errorf("failed to install Gateway API CRDs: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	// Apply our GatewayClass named ck-gateway
	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(GatewayClassChartName), helm.StatePresent, nil); err != nil {
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

	changed, err := helmClient.Apply(ctx, networkManifest.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnly, ciliumValues)
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

func (r reconciler) disableGateway(ctx context.Context, networkManifest *types.FeatureManifest, network types.Network) (types.FeatureStatus, error) {
	ciliumAgentImageTag := networkManifest.GetImage(cilium_network.CiliumAgentImageName).Tag

	helmClient := r.HelmClient()
	snap := r.Snap()

	// Delete our GatewayClass named ck-gateway
	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(GatewayClassChartName), helm.StateDeleted, nil); err != nil {
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

	changed, err := helmClient.Apply(ctx, networkManifest.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), ciliumValues)
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
	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(GatewayChartName), helm.StateDeleted, nil); err != nil {
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

func (r reconciler) getNetworkManifest(ctx context.Context) (*types.FeatureManifest, error) {
	return GetNetworkManifest(ctx, r.State())
}

var GetNetworkManifest = func(ctx context.Context, state state.State) (*types.FeatureManifest, error) {
	var networkManifest *types.FeatureManifest

	if err := state.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		networkManifest, err = database.GetFeatureManifest(ctx, tx, string(features.Network), "1.0.0")
		if err != nil {
			return fmt.Errorf("failed to get network manifest from database: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to perform network manifest transaction request: %w", err)
	}

	return networkManifest, nil
}
