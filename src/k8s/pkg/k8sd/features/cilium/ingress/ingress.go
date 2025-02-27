package ingress

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	IngressDeleteFailedMsgTmpl = "Failed to delete Cilium Ingress, the error was: %v"
	IngressDeployFailedMsgTmpl = "Failed to deploy Cilium Ingress, the error was: %v"

	IngressOptionEnabled                          = "enabled"
	IngressOptionLoadBalancerMode                 = "loadbalancerMode"
	IngressOptionLoadBalancerModeShared           = "shared" // loadbalancerMode: "shared"
	IngressOptionDefaultSecretName                = "defaultSecretName"
	IngressOptionDefaultSecretNamespace           = "defaultSecretNamespace"
	IngressOptionDefaultSecretNamespaceKubeSystem = "kube-system" // defaultSecretNamespace: "kube-system"
	IngressOptionEnableProxyProtocol              = "enableProxyProtocol"
)

// ApplyIngress assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyIngress will enable Cilium's ingress controller when ingress.Enabled is true.
// ApplyIngress will disable Cilium's ingress controller when ingress.Enabled is false.
// ApplyIngress will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyIngress will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyIngress returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyIngress(ctx context.Context, snap snap.Snap, m helm.Client, ingress types.Ingress, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	ciliumAgentImageTag := cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag

	var ciliumValues CiliumValues = map[string]any{}

	if ingress.GetEnabled() {
		if err := ciliumValues.applyDefaultValues(); err != nil {
			err = fmt.Errorf("failed to apply default values: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}

		if err := ciliumValues.applyClusterConfiguration(ingress); err != nil {
			err = fmt.Errorf("failed to apply cluster configuration: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}
	} else {
		if err := ciliumValues.applyDisableConfiguration(); err != nil {
			err = fmt.Errorf("failed to apply disable configuration: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeleteFailedMsgTmpl, err),
			}, err
		}
	}

	changed, err := m.Apply(ctx, cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), ciliumValues)
	if err != nil {
		if network.GetEnabled() {
			err = fmt.Errorf("failed to enable ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to disable ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeleteFailedMsgTmpl, err),
			}, err
		}
	}

	if !changed {
		if ingress.GetEnabled() {
			return types.FeatureStatus{
				Enabled: true,
				Version: ciliumAgentImageTag,
				Message: cilium.EnabledMsg,
			}, nil
		} else {
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: cilium.DisabledMsg,
			}, nil
		}
	}

	if !ingress.GetEnabled() {
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: cilium.DisabledMsg,
		}, nil
	}

	if err := cilium.RolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to apply ingress: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ciliumAgentImageTag,
		Message: cilium.EnabledMsg,
	}, nil
}
