package loadbalancer

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/cilium"
	cilium_network "github.com/canonical/k8s/pkg/k8sd/features/cilium/network"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

const (
	lbEnabledMsgTmpl      = "enabled, %s mode"
	LbDeleteFailedMsgTmpl = "Failed to delete Cilium Load Balancer, the error was: %v"
	LbDeployFailedMsgTmpl = "Failed to deploy Cilium Load Balancer, the error was: %v"
)

// ApplyLoadBalancer assumes that the managed Cilium CNI is already installed on the cluster. It will fail if that is not the case.
// ApplyLoadBalancer will configure Cilium to enable L2 or BGP mode, and deploy necessary CRs for announcing the LoadBalancer external IPs when loadbalancer.Enabled is true.
// ApplyLoadBalancer will disable L2 and BGP on Cilium, and remove any previously created CRs when loadbalancer.Enabled is false.
// ApplyLoadBalancer will rollout restart the Cilium pods in case any Cilium configuration was changed.
// ApplyLoadBalancer will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyLoadBalancer returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyLoadBalancer(ctx context.Context, snap snap.Snap, m helm.Client, loadbalancer types.LoadBalancer, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	ciliumAgentImageTag := cilium_network.FeatureNetwork.GetImage(cilium_network.CiliumAgentImageName).Tag

	if !loadbalancer.GetEnabled() {
		if err := disableLoadBalancer(ctx, snap, m, network); err != nil {
			err = fmt.Errorf("failed to disable LoadBalancer: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ciliumAgentImageTag,
				Message: fmt.Sprintf(LbDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: cilium.DisabledMsg,
		}, nil
	}

	if err := enableLoadBalancer(ctx, snap, m, loadbalancer, network); err != nil {
		err = fmt.Errorf("failed to enable LoadBalancer: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(LbDeployFailedMsgTmpl, err),
		}, err
	}

	switch {
	case loadbalancer.GetBGPMode():
		return types.FeatureStatus{
			Enabled: true,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(lbEnabledMsgTmpl, "BGP"),
		}, nil
	case loadbalancer.GetL2Mode():
		return types.FeatureStatus{
			Enabled: true,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(lbEnabledMsgTmpl, "L2"),
		}, nil
	default:
		return types.FeatureStatus{
			Enabled: true,
			Version: ciliumAgentImageTag,
			Message: fmt.Sprintf(lbEnabledMsgTmpl, "Unknown"),
		}, nil
	}
}

func disableLoadBalancer(ctx context.Context, snap snap.Snap, m helm.Client, network types.Network) error {
	if _, err := m.Apply(ctx, FeatureLoadBalancer.GetChart(LoadbalancerChartName), helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall LoadBalancer manifests: %w", err)
	}

	var ciliumValues CiliumValues = map[string]any{}

	if err := ciliumValues.applyDisableConfiguration(); err != nil {
		return fmt.Errorf("failed to apply disable configuration: %w", err)
	}

	if _, err := m.Apply(ctx, cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), ciliumValues); err != nil {
		return fmt.Errorf("failed to refresh network to apply LoadBalancer configuration: %w", err)
	}
	return nil
}

func enableLoadBalancer(ctx context.Context, snap snap.Snap, m helm.Client, loadbalancer types.LoadBalancer, network types.Network) error {
	var ciliumValues CiliumValues = map[string]any{}

	if err := ciliumValues.applyDefaultValues(); err != nil {
		return fmt.Errorf("failed to apply default values: %w", err)
	}

	if err := ciliumValues.applyClusterConfiguration(loadbalancer); err != nil {
		return fmt.Errorf("failed to apply cluster configuration: %w", err)
	}

	changed, err := m.Apply(ctx, cilium_network.FeatureNetwork.GetChart(cilium_network.CiliumChartName), helm.StateUpgradeOnlyOrDeleted(network.GetEnabled()), ciliumValues)
	if err != nil {
		return fmt.Errorf("failed to update Cilium configuration for LoadBalancer: %w", err)
	}

	if err := waitForRequiredLoadBalancerCRDs(ctx, snap, loadbalancer.GetBGPMode()); err != nil {
		return fmt.Errorf("failed to wait for required Cilium CRDs to be available: %w", err)
	}

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		return fmt.Errorf("failed to apply default values: %w", err)
	}

	if err := values.applyClusterConfiguration(loadbalancer); err != nil {
		return fmt.Errorf("failed to apply cluster configuration: %w", err)
	}

	if _, err := m.Apply(ctx, FeatureLoadBalancer.GetChart(LoadbalancerChartName), helm.StatePresent, values); err != nil {
		return fmt.Errorf("failed to apply LoadBalancer configuration: %w", err)
	}

	if !changed {
		return nil
	}

	if err := cilium.RolloutRestartCilium(ctx, snap, 3); err != nil {
		return fmt.Errorf("failed to rollout restart cilium to apply LoadBalancer configuration: %w", err)
	}
	return nil
}

func waitForRequiredLoadBalancerCRDs(ctx context.Context, snap snap.Snap, bgpMode bool) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return control.WaitUntilReady(ctx, func() (bool, error) {
		resources, err := client.ListResourcesForGroupVersion("cilium.io/v2alpha1")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}

		requiredCRDs := map[string]struct{}{
			"ciliuml2announcementpolicies": {},
			"ciliumloadbalancerippools":    {},
		}
		if bgpMode {
			requiredCRDs["ciliumbgppeeringpolicies"] = struct{}{}
		}
		requiredCount := len(requiredCRDs)
		for _, resource := range resources.APIResources {
			if _, ok := requiredCRDs[resource.Name]; ok {
				requiredCount--
			}
		}
		return requiredCount == 0, nil
	})
}
