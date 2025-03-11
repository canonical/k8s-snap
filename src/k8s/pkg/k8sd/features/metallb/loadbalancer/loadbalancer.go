package loadbalancer

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/metallb"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/control"
)

const (
	deleteFailedMsgTmpl = "Failed to delete MetalLB, the error was: %v"
	deployFailedMsgTmpl = "Failed to deploy MetalLB, the error was: %v"
)

// ApplyLoadBalancer will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyLoadBalancer returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func (r reconciler) Reconcile(ctx context.Context, cfg types.ClusterConfig) (types.FeatureStatus, error) {
	metalLBControllerImage := r.Manifest().GetImage(MetalLBControllerImageName)

	loadbalancer := cfg.LoadBalancer

	if !loadbalancer.GetEnabled() {
		if err := r.disableLoadBalancer(ctx); err != nil {
			err = fmt.Errorf("failed to disable LoadBalancer: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: metalLBControllerImage.Tag,
				Message: fmt.Sprintf(deleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: metalLBControllerImage.Tag,
			Message: metallb.DisabledMsg,
		}, nil
	}

	if err := r.enableLoadBalancer(ctx, loadbalancer); err != nil {
		err = fmt.Errorf("failed to enable LoadBalancer: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: metalLBControllerImage.Tag,
			Message: fmt.Sprintf(deployFailedMsgTmpl, err),
		}, err
	}

	switch {
	case loadbalancer.GetBGPMode():
		return types.FeatureStatus{
			Enabled: true,
			Version: metalLBControllerImage.Tag,
			Message: fmt.Sprintf(metallb.EnabledMsgTmpl, "BGP"),
		}, nil
	case loadbalancer.GetL2Mode():
		return types.FeatureStatus{
			Enabled: true,
			Version: metalLBControllerImage.Tag,
			Message: fmt.Sprintf(metallb.EnabledMsgTmpl, "L2"),
		}, nil
	default:
		return types.FeatureStatus{
			Enabled: true,
			Version: metalLBControllerImage.Tag,
			Message: fmt.Sprintf(metallb.EnabledMsgTmpl, "Unknown"),
		}, nil
	}
}

func (r reconciler) disableLoadBalancer(ctx context.Context) error {
	helmClient := r.HelmClient()

	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(LoadBalancerChartName), helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall MetalLB LoadBalancer chart: %w", err)
	}

	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(MetalLBChartName), helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall MetalLB chart: %w", err)
	}
	return nil
}

func (r reconciler) enableLoadBalancer(ctx context.Context, loadbalancer types.LoadBalancer) error {
	helmClient := r.HelmClient()

	var metalLBValues MetalLBValues = map[string]any{}

	if err := metalLBValues.applyDefaultValues(); err != nil {
		return fmt.Errorf("failed to apply default values: %w", err)
	}

	if err := metalLBValues.ApplyImageOverrides(r.Manifest()); err != nil {
		return fmt.Errorf("failed to apply image overrides: %w", err)
	}

	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(MetalLBChartName), helm.StatePresent, metalLBValues); err != nil {
		return fmt.Errorf("failed to apply MetalLB configuration: %w", err)
	}

	if err := r.waitForRequiredLoadBalancerCRDs(ctx, loadbalancer.GetBGPMode()); err != nil {
		return fmt.Errorf("failed to wait for required MetalLB CRDs: %w", err)
	}

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		return fmt.Errorf("failed to apply default values: %w", err)
	}

	if err := values.applyClusterConfiguration(loadbalancer); err != nil {
		return fmt.Errorf("failed to apply cluster configuration: %w", err)
	}

	if _, err := helmClient.Apply(ctx, r.Manifest().GetChart(LoadBalancerChartName), helm.StatePresent, values); err != nil {
		return fmt.Errorf("failed to apply MetalLB LoadBalancer configuration: %w", err)
	}

	return nil
}

func (r reconciler) waitForRequiredLoadBalancerCRDs(ctx context.Context, bgpMode bool) error {
	client, err := r.Snap().KubernetesClient("")
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
