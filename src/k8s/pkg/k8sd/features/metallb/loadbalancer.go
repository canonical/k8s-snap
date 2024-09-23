package metallb

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
)

const (
	enabledMsgTmpl      = "enabled, %s mode"
	DisabledMsg         = "disabled"
	deleteFailedMsgTmpl = "Failed to delete MetalLB, the error was: %v"
	deployFailedMsgTmpl = "Failed to deploy MetalLB, the error was: %v"
)

// ApplyLoadBalancer will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyLoadBalancer returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	if !loadbalancer.GetEnabled() {
		if err := disableLoadBalancer(ctx, snap, network); err != nil {
			err = fmt.Errorf("failed to disable LoadBalancer: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ControllerImageTag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: ControllerImageTag,
			Message: DisabledMsg,
		}, nil
	}

	if err := enableLoadBalancer(ctx, snap, loadbalancer, network); err != nil {
		err = fmt.Errorf("failed to enable LoadBalancer: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ControllerImageTag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	if loadbalancer.GetBGPMode() {
		return types.FeatureStatus{
			Enabled: true,
			Version: ControllerImageTag,
			Message: fmt.Sprintf(enabledMsgTmpl, "BGP"),
		}, nil
	} else if loadbalancer.GetL2Mode() {
		return types.FeatureStatus{
			Enabled: true,
			Version: ControllerImageTag,
			Message: fmt.Sprintf(enabledMsgTmpl, "L2"),
		}, nil
	} else {
		return types.FeatureStatus{
			Enabled: true,
			Version: ControllerImageTag,
			Message: fmt.Sprintf(enabledMsgTmpl, "Unknown"),
		}, nil
	}
}

func disableLoadBalancer(ctx context.Context, snap snap.Snap, network types.Network) error {
	m := snap.HelmClient()

	if _, err := m.Apply(ctx, ChartMetalLBLoadBalancer, helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall MetalLB LoadBalancer chart: %w", err)
	}

	if _, err := m.Apply(ctx, ChartMetalLB, helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall MetalLB chart: %w", err)
	}
	return nil
}

func enableLoadBalancer(ctx context.Context, snap snap.Snap, loadbalancer types.LoadBalancer, network types.Network) error {
	m := snap.HelmClient()

	metalLBValues := map[string]any{
		"controller": map[string]any{
			"image": map[string]any{
				"repository": controllerImageRepo,
				"tag":        ControllerImageTag,
			},
		},
		"speaker": map[string]any{
			"image": map[string]any{
				"repository": speakerImageRepo,
				"tag":        speakerImageTag,
			},
			// TODO(neoaggelos): make frr enable/disable configurable through an annotation
			// We keep it disabled by default
			"frr": map[string]any{
				"enabled": false,
				"image": map[string]any{
					"repository": frrImageRepo,
					"tag":        frrImageTag,
				},
			},
		},
	}
	if _, err := m.Apply(ctx, ChartMetalLB, helm.StatePresent, metalLBValues); err != nil {
		return fmt.Errorf("failed to apply MetalLB configuration: %w", err)
	}

	if err := waitForRequiredLoadBalancerCRDs(ctx, snap, loadbalancer.GetBGPMode()); err != nil {
		return fmt.Errorf("failed to wait for required MetalLB CRDs: %w", err)
	}

	cidrs := []map[string]any{}
	for _, cidr := range loadbalancer.GetCIDRs() {
		cidrs = append(cidrs, map[string]any{"cidr": cidr})
	}
	for _, ipRange := range loadbalancer.GetIPRanges() {
		cidrs = append(cidrs, map[string]any{"start": ipRange.Start, "stop": ipRange.Stop})
	}

	values := map[string]any{
		"driver": "metallb",
		"l2": map[string]any{
			"enabled":    loadbalancer.GetL2Mode(),
			"interfaces": loadbalancer.GetL2Interfaces(),
		},
		"ipPool": map[string]any{
			"cidrs": cidrs,
		},
		"bgp": map[string]any{
			"enabled":  loadbalancer.GetBGPMode(),
			"localASN": loadbalancer.GetBGPLocalASN(),
			"neighbors": []map[string]any{
				{
					"peerAddress": loadbalancer.GetBGPPeerAddress(),
					"peerASN":     loadbalancer.GetBGPPeerASN(),
					"peerPort":    loadbalancer.GetBGPPeerPort(),
				},
			},
		},
	}

	if _, err := m.Apply(ctx, ChartMetalLBLoadBalancer, helm.StatePresent, values); err != nil {
		return fmt.Errorf("failed to apply MetalLB LoadBalancer configuration: %w", err)
	}

	return nil
}

func waitForRequiredLoadBalancerCRDs(ctx context.Context, snap snap.Snap, bgpMode bool) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return control.WaitUntilReady(ctx, func() (bool, error) {
		resourcesv1beta1, err := client.ListResourcesForGroupVersion("metallb.io/v1beta1")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}
		resourcesv1beta2, err := client.ListResourcesForGroupVersion("metallb.io/v1beta2")
		if err != nil {
			// This error is expected if the group version is not yet deployed.
			return false, nil
		}

		requiredCRDs := map[string]struct{}{
			"metallb.io/v1beta1:ipaddresspools":   {},
			"metallb.io/v1beta1:l2advertisements": {},
		}
		if bgpMode {
			requiredCRDs["metallb.io/v1beta2:bgppeers"] = struct{}{}
			requiredCRDs["metallb.io/v1beta1:bgpadvertisements"] = struct{}{}
		}

		requiredCount := len(requiredCRDs)

		for _, resource := range resourcesv1beta1.APIResources {
			if _, ok := requiredCRDs[fmt.Sprintf("metallb.io/v1beta1:%s", resource.Name)]; ok {
				requiredCount--
			}
		}

		for _, resource := range resourcesv1beta2.APIResources {
			if _, ok := requiredCRDs[fmt.Sprintf("metallb.io/v1beta2:%s", resource.Name)]; ok {
				requiredCount--
			}
		}

		return requiredCount == 0, nil
	})
}
