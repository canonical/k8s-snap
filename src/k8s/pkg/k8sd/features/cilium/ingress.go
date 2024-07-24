package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	ingressDeleteFailedMsgTmpl = "Failed to delete Cilium Ingress, the error was: %v"
	ingressDeployFailedMsgTmpl = "Failed to deploy Cilium Ingress, the error was: %v"
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
	status := types.FeatureStatus{
		Version: ciliumAgentImageTag,
		Enabled: ingress.GetEnabled(),
	}
	m := snap.HelmClient()

	var values map[string]any
	if ingress.GetEnabled() {
		values = map[string]any{
			"ingressController": map[string]any{
				"enabled":                true,
				"loadbalancerMode":       "shared",
				"defaultSecretNamespace": "kube-system",
				"defaultTLSSecret":       ingress.GetDefaultTLSSecret(),
				"enableProxyProtocol":    ingress.GetEnableProxyProtocol(),
			},
		}
	} else {
		values = map[string]any{
			"ingressController": map[string]any{
				"enabled":                false,
				"defaultSecretNamespace": "",
				"defaultSecretName":      "",
				"enableProxyProtocol":    false,
			},
		}
	}

	changed, err := m.Apply(ctx, chartCilium, helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), values)
	if err != nil {
		if network.GetEnabled() {
			enableErr := fmt.Errorf("failed to enable ingress: %w", err)
			status.Message = fmt.Sprintf(ingressDeployFailedMsgTmpl, enableErr)
			status.Enabled = false
			return status, enableErr
		} else {
			disableErr := fmt.Errorf("failed to disable ingress: %w", err)
			status.Message = fmt.Sprintf(ingressDeleteFailedMsgTmpl, disableErr)
			return status, disableErr
		}
	}

	if !changed {
		if ingress.GetEnabled() {
			status.Message = enabledMsg
			return status, nil
		} else {
			status.Message = disabledMsg
			status.Version = ""
			return status, nil
		}
	}

	if !ingress.GetEnabled() {
		status.Message = disabledMsg
		status.Version = ""
		return status, nil
	}

	if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
		restartErr := fmt.Errorf("failed to rollout restart cilium to apply ingress: %w", err)
		status.Message = fmt.Sprintf(ingressDeployFailedMsgTmpl, restartErr)
		status.Enabled = false
		return status, restartErr
	}

	status.Message = enabledMsg
	return status, nil
}
