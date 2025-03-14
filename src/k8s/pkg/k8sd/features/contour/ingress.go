package contour

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
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
func ApplyIngress(ctx context.Context, snap snap.Snap, ingress types.Ingress, _ types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	if !ingress.GetEnabled() {
		if _, err := m.Apply(ctx, chartContour, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ContourIngressContourImageTag,
				Message: fmt.Sprintf(IngressDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: ContourIngressContourImageTag,
			Message: DisabledMsg,
		}, nil
	}

	// Apply common contour CRDS, these are shared with gateway
	if err := applyCommonContourCRDS(ctx, snap, true); err != nil {
		err = fmt.Errorf("failed to apply common contour CRDS: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ContourIngressContourImageTag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	if err := waitForRequiredContourCommonCRDs(ctx, snap); err != nil {
		err = fmt.Errorf("failed to wait for required contour common CRDs to be available: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ContourIngressContourImageTag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	var values map[string]any
	values = map[string]any{
		"envoy-service-namespace": "projectcontour",
		"envoy-service-name":      "envoy",
		"envoy": map[string]any{
			"image": map[string]any{
				"registry":   "",
				"repository": ContourIngressEnvoyImageRepo,
				"tag":        ContourIngressEnvoyImageTag,
			},
		},
		"contour": map[string]any{
			"manageCRDs": false,
			"ingressClass": map[string]any{
				"name":    "ck-ingress",
				"create":  true,
				"default": true,
			},
			"image": map[string]any{
				"registry":   "",
				"repository": ContourIngressContourImageRepo,
				"tag":        ContourIngressContourImageTag,
			},
		},
	}

	if ingress.GetEnableProxyProtocol() {
		contour, ok := values["contour"].(map[string]any)
		if !ok {
			err := fmt.Errorf("unexpected type for contour values")
			return types.FeatureStatus{
				Enabled: false,
				Version: ContourIngressContourImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}
		contour["extraArgs"] = []string{"--use-proxy-protocol"}
	}

	changed, err := m.Apply(ctx, chartContour, helm.StatePresent, values)
	if err != nil {
		err = fmt.Errorf("failed to enable ingress: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ContourIngressContourImageTag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err
	}

	if changed {
		if err := rolloutRestartContour(ctx, snap, 3); err != nil {
			err = fmt.Errorf("failed to rollout restart contour to apply ingress: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ContourIngressContourImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}
	}

	// Install the delegation resource for the default TLS secret.
	// The default TLS secret is created by the user
	// and gets set via k8s set defaultTLSSecret=bananas.
	if ingress.GetDefaultTLSSecret() != "" {
		values = map[string]any{
			"defaultTLSSecret": ingress.GetDefaultTLSSecret(),
		}
		if _, err := m.Apply(ctx, chartDefaultTLS, helm.StatePresent, values); err != nil {
			err = fmt.Errorf("failed to install the delegation resource for default TLS secret: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: ContourIngressContourImageTag,
				Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: true,
			Version: ContourIngressContourImageTag,
			Message: EnabledMsg,
		}, nil
	}

	if _, err := m.Apply(ctx, chartDefaultTLS, helm.StateDeleted, nil); err != nil {
		err = fmt.Errorf("failed to uninstall the delegation resource for default TLS secret: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: ContourIngressContourImageTag,
			Message: fmt.Sprintf(IngressDeployFailedMsgTmpl, err),
		}, err

	}

	return types.FeatureStatus{
		Enabled: true,
		Version: ContourIngressContourImageTag,
		Message: EnabledMsg,
	}, nil
}

// applyCommonContourCRDS will install the common contour CRDS when enabled is true.
// These CRDS are shared between the contour ingress and the gateway feature.
func applyCommonContourCRDS(ctx context.Context, snap snap.Snap, enabled bool) error {
	m := snap.HelmClient()
	if enabled {
		if _, err := m.Apply(ctx, chartCommonContourCRDS, helm.StatePresent, nil); err != nil {
			return fmt.Errorf("failed to install common CRDS: %w", err)
		}
		return nil
	}

	if _, err := m.Apply(ctx, chartCommonContourCRDS, helm.StateDeleted, nil); err != nil {
		return fmt.Errorf("failed to uninstall common CRDS: %w", err)
	}

	return nil
}

// rolloutRestartContour will rollout restart the Contour pods in case any Contour configuration was changed.
func rolloutRestartContour(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "ck-ingress-contour-contour", "projectcontour"); err != nil {
			return fmt.Errorf("failed to restart contour deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart contour deployment after %d attempts: %w", attempts, err)
	}

	return nil
}
