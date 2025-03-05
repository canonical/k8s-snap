package network

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
)

const (
	NetworkDeleteFailedMsgTmpl = "Failed to delete Cilium Network, the error was: %v"
	NetworkDeployFailedMsgTmpl = "Failed to deploy Cilium Network, the error was: %v"
)

// required for unittests.
var (
	GetMountPath            = utils.GetMountPath
	GetMountPropagationType = utils.GetMountPropagationType
)

// ApplyNetwork will deploy Cilium when network.Enabled is true.
// ApplyNetwork will remove Cilium when network.Enabled is false.
// ApplyNetwork requires that bpf and cgroups2 are already mounted and available when running under strict snap confinement. If they are not, it will fail (since Cilium will not have the required permissions to mount them).
// ApplyNetwork requires that `/sys` is mounted as a shared mount when running under classic snap confinement. This is to ensure that Cilium will be able to automatically mount bpf and cgroups2 on the pods.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func (r reconciler) Reconcile(ctx context.Context, cfg types.ClusterConfig) (types.FeatureStatus, error) {
	ciliumAgentImage := r.Manifest().GetImage(CiliumAgentImageName)

	helmClient := r.HelmClient()
	snap := r.Snap()

	network := cfg.Network
	apiserver := cfg.APIServer
	annotations := cfg.Annotations

	if !network.GetEnabled() {
		if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(CiliumChartName), helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall network: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImage.Tag,
				Message: fmt.Sprintf(NetworkDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImage.Tag,
			Message: cilium.DisabledMsg,
		}, nil
	}

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		err = fmt.Errorf("failed to apply default values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImage.Tag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyImageOverrides(r.Manifest()); err != nil {
		err = fmt.Errorf("failed to calculate image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImage.Tag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if snap.Strict() {
		if err := values.ApplyStrictOverrides(); err != nil {
			err = fmt.Errorf("failed to calculate strict overrides: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImage.Tag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
			}, err
		}
	}

	if err := values.applyClusterConfiguration(ctx, r.State(), apiserver, network); err != nil {
		err = fmt.Errorf("failed to calculate cluster config values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImage.Tag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyAnnotations(annotations); err != nil {
		err = fmt.Errorf("failed to calculate annotation overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImage.Tag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if !snap.Strict() {
		if err := r.verifyMountPropagation(ctx); err != nil {
			err = fmt.Errorf("failed to check mount propagation: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImage.Tag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
			}, err
		}
	}

	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(CiliumChartName), helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to enable network: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImage.Tag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ciliumAgentImage.Tag,
		Message: cilium.EnabledMsg,
	}, nil
}
