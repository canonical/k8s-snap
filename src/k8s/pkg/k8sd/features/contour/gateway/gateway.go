package gateway

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features/contour"
	contour_ingress "github.com/canonical/k8s/pkg/k8sd/features/contour/ingress"
	"github.com/canonical/k8s/pkg/k8sd/types"
)

const (
	GatewayDeployFailedMsgTmpl = "Failed to deploy Contour Gateway, the error was: %v"
	GatewayDeleteFailedMsgTmpl = "Failed to delete Contour Gateway, the error was: %v"
)

// ApplyGateway will install a helm chart for contour-gateway-provisioner on the cluster when gateway.Enabled is true.
// ApplyGateway will uninstall the helm chart for contour-gateway-provisioner from the cluster when gateway.Enabled is false.
// ApplyGateway will apply common contour CRDS, these are shared with ingress.
// ApplyGateway will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyGateway returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func (r GatewayReconciler) ApplyGateway(ctx context.Context, gateway types.Gateway, network types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	contourGatewayProvisionerContourImage := FeatureGateway.GetImage(ContourGatewayProvisionerContourImageName)

	helmClient := r.HelmClient()
	snap := r.Snap()

	if !gateway.GetEnabled() {
		if _, err := helmClient.Apply(ctx, FeatureGateway.GetChart(ChartGatewayName), helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall the contour gateway chart: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: contourGatewayProvisionerContourImage.Tag,
				Message: fmt.Sprintf(GatewayDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: contourGatewayProvisionerContourImage.Tag,
			Message: contour.DisabledMsg,
		}, nil
	}

	// Apply common contour CRDS, these are shared with ingress
	if err := contour_ingress.ApplyCommonContourCRDS(ctx, snap, helmClient, true); err != nil {
		err = fmt.Errorf("failed to apply common contour CRDS: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourGatewayProvisionerContourImage.Tag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	if err := contour.WaitForRequiredContourCommonCRDs(ctx, snap); err != nil {
		err = fmt.Errorf("failed to wait for required contour common CRDs to be available: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourGatewayProvisionerContourImage.Tag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	var values Values = map[string]any{}

	if err := values.ApplyImageOverrides(); err != nil {
		err = fmt.Errorf("failed to apply image overrides: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourGatewayProvisionerContourImage.Tag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	if _, err := helmClient.Apply(ctx, FeatureGateway.GetChart(ChartGatewayName), helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to install the contour gateway chart: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: contourGatewayProvisionerContourImage.Tag,
			Message: fmt.Sprintf(GatewayDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: contourGatewayProvisionerContourImage.Tag,
		Message: contour.EnabledMsg,
	}, nil
}
