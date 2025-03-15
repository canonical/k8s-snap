package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
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
func ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()
	var values map[string]any
	if ingress.GetEnabled() {
		values = map[string]any{
			"ingressController": map[string]any{
				IngressOptionEnabled:                true,
				IngressOptionLoadBalancerMode:       IngressOptionLoadBalancerModeShared,
				IngressOptionDefaultSecretNamespace: IngressOptionDefaultSecretNamespaceKubeSystem,
				IngressOptionDefaultSecretName:      ingress.GetDefaultTLSSecret(),
				IngressOptionEnableProxyProtocol:    ingress.GetEnableProxyProtocol(),
			},
		}
	} else {
		values = map[string]any{
			"ingressController": map[string]any{
				IngressOptionEnabled:                false,
				IngressOptionLoadBalancerMode:       "",
				IngressOptionDefaultSecretNamespace: "",
				IngressOptionDefaultSecretName:      "",
				IngressOptionEnableProxyProtocol:    false,
			},
		}
	}

	changed, err := m.Apply(ctx, ChartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), values)
	if err != nil {
		if network.GetEnabled() {
			err = fmt.Errorf("failed to enable ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		} else {
			err = fmt.Errorf("failed to disable ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(IngressDeleteFailedMsgTmpl, err),
			}, err
		}
	}

	if !changed {
		if ingress.GetEnabled() {
			return types.FeatureStatus{
				Enabled: true,
				Version: CiliumAgentImageTag,
				Message: EnabledMsg,
			}, nil
		} else {
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: DisabledMsg,
			}, nil
		}
	}

	if !ingress.GetEnabled() {
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: DisabledMsg,
		}, nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		err = fmt.Errorf("failed to rollout restart cilium to apply ingress: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: CiliumAgentImageTag,
		Message: EnabledMsg,
	}, nil
}
