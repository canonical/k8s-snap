package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/features"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
	"github.com/canonical/microcluster/v2/state"
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

const NETWORK_VERSION = "v1.0.0"

// ApplyNetwork will deploy Cilium when network.Enabled is true.
// ApplyNetwork will remove Cilium when network.Enabled is false.
// ApplyNetwork requires that bpf and cgroups2 are already mounted and available when running under strict snap confinement. If they are not, it will fail (since Cilium will not have the required permissions to mount them).
// ApplyNetwork requires that `/sys` is mounted as a shared mount when running under classic snap confinement. This is to ensure that Cilium will be able to automatically mount bpf and cgroups2 on the pods.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyNetwork(ctx context.Context, s state.State, snap snap.Snap, apiserver types.APIServer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	if !network.GetEnabled() {
		if _, err := m.Apply(ctx, features.Network, NETWORK_VERSION, ChartCilium, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall network: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(NetworkDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: DisabledMsg,
		}, nil
	}

	values := networkValues{}

	if err := values.applyDefaults(); err != nil {
		err := fmt.Errorf("failed to apply defaults: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyImages(); err != nil {
		err := fmt.Errorf("failed to apply images: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if snap.Strict() {
		if err := values.applyStrict(); err != nil {
			err := fmt.Errorf("failed to apply strict configuration: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
			}, err
		}
	}

	if err := values.applyClusterConfig(ctx, s, apiserver, network); err != nil {
		err := fmt.Errorf("failed to apply cluster config: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if err := values.applyAnnotations(annotations); err != nil {
		err := fmt.Errorf("failed to apply annotations: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	if !snap.Strict() {
		pt, err := GetMountPropagationType("/sys")
		if err != nil {
			err = fmt.Errorf("failed to get mount propagation type for /sys: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
			}, err
		}
		if pt == utils.MountPropagationPrivate {
			onLXD, err := snap.OnLXD(ctx)
			if err != nil {
				logger := log.FromContext(ctx)
				logger.Error(err, "Failed to check if running on LXD")
			}
			if onLXD {
				err := fmt.Errorf("/sys is not a shared mount on the LXD container, this might be resolved by updating LXD on the host to version 5.0.2 or newer")
				return types.FeatureStatus{
					Enabled: false,
					Version: CiliumAgentImageTag,
					Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
				}, err
			}

			err = fmt.Errorf("/sys is not a shared mount")
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
			}, err
		}
	}

	if _, err := m.Apply(ctx, features.Network, NETWORK_VERSION, ChartCilium, helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to enable network: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	return types.FeatureStatus{
		Enabled: true,
		Version: CiliumAgentImageTag,
		Message: EnabledMsg,
	}, nil
}

func rolloutRestartCilium(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "cilium-operator", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment after %d attempts: %w", attempts, err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDaemonset(ctx, "cilium", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium daemonset after %d attempts: %w", attempts, err)
	}

	return nil
}
