package cilium

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
)

const (
	networkDeleteFailedMsgTmpl = "Failed to delete Cilium Network, the error was: %v"
	networkDeployFailedMsgTmpl = "Failed to deploy Cilium Network, the error was: %v"
)

// required for unittests.
var (
	getMountPath            = utils.GetMountPath
	getMountPropagationType = utils.GetMountPropagationType
)

// ApplyNetwork will deploy Cilium when network.Enabled is true.
// ApplyNetwork will remove Cilium when network.Enabled is false.
// ApplyNetwork requires that bpf and cgroups2 are already mounted and available when running under strict snap confinement. If they are not, it will fail (since Cilium will not have the required permissions to mount them).
// ApplyNetwork requires that `/sys` is mounted as a shared mount when running under classic snap confinement. This is to ensure that Cilium will be able to automatically mount bpf and cgroups2 on the pods.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyNetwork(ctx context.Context, snap snap.Snap, localhostAddress string, apiserver types.APIServer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	if !network.GetEnabled() {
		if _, err := m.Apply(ctx, ChartCilium, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall network: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(networkDeleteFailedMsgTmpl, err),
			}, err
		}
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: DisabledMsg,
		}, nil
	}

	config, err := internalConfig(annotations)
	if err != nil {
		err = fmt.Errorf("failed to parse annotations: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
		}, err
	}

	ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	if err != nil {
		err = fmt.Errorf("invalid kube-proxy --cluster-cidr value: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
		}, err
	}

	ciliumNodePortValues := map[string]any{
		"enabled": true,
		// kube-proxy also binds to the same port for health checks so we need to disable it
		"enableHealthCheck": false,
	}

	if config.directRoutingDevice != "" {
		ciliumNodePortValues["directRoutingDevice"] = config.directRoutingDevice
	}

	bpfValues := map[string]any{}
	if config.vlanBPFBypass != nil {
		bpfValues["vlanBypass"] = config.vlanBPFBypass
	}

	values := map[string]any{
		"bpf": bpfValues,
		"image": map[string]any{
			"repository": ciliumAgentImageRepo,
			"tag":        CiliumAgentImageTag,
			"useDigest":  false,
		},
		"socketLB": map[string]any{
			"enabled": true,
		},
		"cni": map[string]any{
			"confPath": "/etc/cni/net.d",
			"binPath":  "/opt/cni/bin",
		},
		"operator": map[string]any{
			"replicas": 1,
			"image": map[string]any{
				"repository": ciliumOperatorImageRepo,
				"tag":        ciliumOperatorImageTag,
				"useDigest":  false,
			},
		},
		"ipv4": map[string]any{
			"enabled": ipv4CIDR != "",
		},
		"ipv6": map[string]any{
			"enabled": ipv6CIDR != "",
		},
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": ipv4CIDR,
				"clusterPoolIPv6PodCIDRList": ipv6CIDR,
			},
		},
		// https://docs.cilium.io/en/v1.15/network/kubernetes/kubeproxy-free/#kube-proxy-hybrid-modes
		"nodePort":                 ciliumNodePortValues,
		"disableEnvoyVersionCheck": true,
		// socketLB requires an endpoint to the apiserver that's not managed by the kube-proxy
		// so we point to the localhost:secureport to talk to either the kube-apiserver or the kube-apiserver-proxy
		"k8sServiceHost": strings.Trim(localhostAddress, "[]"), // Cilium already adds the brackets for ipv6 addresses, so we need to remove them
		"k8sServicePort": apiserver.GetSecurePort(),
		// This flag enables the runtime device detection which is set to true by default in Cilium 1.16+
		"enableRuntimeDeviceDetection": true,
	}

	if config.devices != "" {
		values["devices"] = config.devices
	}

	if snap.Strict() {
		bpfMnt, err := getMountPath("bpf")
		if err != nil {
			err = fmt.Errorf("failed to get bpf mount path: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
			}, err
		}

		cgrMnt, err := getMountPath("cgroup2")
		if err != nil {
			err = fmt.Errorf("failed to get cgroup2 mount path: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
			}, err
		}

		values["bpf"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"root": bpfMnt,
		}
		values["cgroup"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"hostRoot": cgrMnt,
		}
	} else {
		pt, err := getMountPropagationType("/sys")
		if err != nil {
			err = fmt.Errorf("failed to get mount propagation type for /sys: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
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
					Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
				}, err
			}

			err = fmt.Errorf("/sys is not a shared mount")
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
			}, err
		}
	}

	if _, err := m.Apply(ctx, ChartCilium, helm.StatePresent, values); err != nil {
		err = fmt.Errorf("failed to enable network: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(networkDeployFailedMsgTmpl, err),
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
