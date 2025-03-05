package ingress

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/contour"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
)

const (
	IngressDeleteFailedMsgTmpl = "Failed to delete Contour Ingress, the error was: %v"
	IngressDeployFailedMsgTmpl = "Failed to deploy Contour Ingress, the error was: %v"
)

// ApplyIngress will install the contour helm chart when ingress.Enabled is true.
// ApplyIngress will uninstall the contour helm chart when ingress.Disabled is false.
// ApplyIngress will rollout restart the Contour pods in case any Contour configuration was changed.
// ApplyIngress will install a delegation resource via helm chart
// for the default TLS secret if ingress.DefaultTLSSecret is set.
// ApplyIngress will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyIngress returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
// Contour CRDS are applied through a ck-contour common chart (Overlap with gateway).
func (r IngressReconciler) ApplyIngress(ctx context.Context, ingress types.Ingress, _ types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	contourIngressContourImage := FeatureIngress.GetImage(ContourIngressContourImageName)

	helmClient := r.HelmClient()
	snap := r.Snap()

	if !ingress.GetEnabled() {
		if _, err := helmClient.Apply(ctx, FeatureIngress.GetChart(ChartContourName), helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: contourIngressContourImage.Tag,
				Message: fmt.Sprintf(IngressDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: contour.DisabledMsg,
		}, nil
	}

	// Apply common contour CRDS, these are shared with gateway
	if err := ApplyCommonContourCRDS(ctx, snap, helmClient, true); err != nil {
		err = fmt.Errorf("failed to apply common contour CRDS: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	if err := contour.WaitForRequiredContourCommonCRDs(ctx, snap); err != nil {
		err = fmt.Errorf("failed to wait for required contour common CRDs to be available: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	var values Values = map[string]any{}

	if err := values.applyDefaultValues(); err != nil {
		err = fmt.Errorf("failed to apply default values: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	if err := values.ApplyImageOverrides(); err != nil {
		err = fmt.Errorf("failed to apply image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyClusterConfiguration(ingress); err != nil {
		err = fmt.Errorf("failed to apply cluster configuration: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	changed, err := helmClient.Apply(ctx, FeatureIngress.GetChart(ChartContourName), helm.StatePresent, values)
	if err != nil {
		err = fmt.Errorf("failed to enable ingress: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	if changed {
		if err := contour.RolloutRestartContour(ctx, snap, 3); err != nil {
			err = fmt.Errorf("failed to rollout restart contour to apply ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: contourIngressContourImage.Tag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}
	}

	// Install the delegation resource for the default TLS secret.
	// The default TLS secret is created by the user
	// and gets set via k8s set defaultTLSSecret=bananas.
	if ingress.GetDefaultTLSSecret() != "" {
		var tlsValues TLSValues = map[string]any{}

		if err := tlsValues.applyClusterConfiguration(ingress); err != nil {
			err = fmt.Errorf("failed to apply cluster configuration: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: contourIngressContourImage.Tag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}

		if _, err := helmClient.Apply(ctx, FeatureIngress.GetChart(ChartDefaultTLSName), helm.StatePresent, tlsValues); err != nil {
			err = fmt.Errorf("failed to install the delegation resource for default TLS secret: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: contourIngressContourImage.Tag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: true,
			Version: contourIngressContourImage.Tag,
			Message: contour.EnabledMsg,
		}, nil
	}

	if _, err := helmClient.Apply(ctx, FeatureIngress.GetChart(ChartDefaultTLSName), helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to uninstall the delegation resource for default TLS secret: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourIngressContourImage.Tag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err

	}

	return types.FeatureStatus{
		Enabled: true,
		Version: contourIngressContourImage.Tag,
		Message: contour.EnabledMsg,
	}, nil
}

// applyCommonContourCRDS will install the common contour CRDS when enabled is true.
// These CRDS are shared between the contour ingress and the gateway feature.
func ApplyCommonContourCRDS(ctx context.Context, snap snap.Snap, helmClient helm.Client, enabled bool) error {
	if enabled {
		if _, err := helmClient.Apply(ctx, FeatureIngress.GetChart(ChartCommonContourCRDSName), helm.StatePresent, nil); err != nil {
			return fmt.Errorf("failed to install common CRDS: %w", err)
		}
		return nil
	}

	if _, err := helmClient.Apply(ctx, FeatureIngress.GetChart(ChartCommonContourCRDSName), helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall common CRDS: %w", err)
	}

	return nil
}
