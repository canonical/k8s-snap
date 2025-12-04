package cilium

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/canonical/k8s/pkg/client/helm"
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

// Cilium uses vxlan encapsulation protocol by default. Since we are using the default
// cilium tunnel encapsulation protocol, we have to make sure that Cilium's vxlan is the
// only interface using the default vxlan port. Otherwise, Cilium might conflict with
// other tools such as fan-netwotking, which use the same vxlan destination port.
func checkAndSanitizeCiliumVXLAN(port int) error {
	vxlanDevices, err := utils.ListVXLANInterfaces()
	if err != nil {
		return fmt.Errorf("listing vxlan interfaces failed: %w", err)
	}

	for _, vxlanDevice := range vxlanDevices {
		if vxlanDevice.Port == nil {
			// This vxlan interface does not have a port set
			// or it was not included in the output of `ip -d -j link list type vxlan`.
			continue
		}

		devicePort := *vxlanDevice.Port

		if devicePort == port && vxlanDevice.Name != ciliumVXLANDeviceName {
			return fmt.Errorf("interface %s uses the same destination port as cilium. Please consider changing the Cilium tunnel port", vxlanDevice.Name)
		}

		// Note(Reza): Currently Cilium tries to bring up the vxlan interface before applying
		// any configuration changes. If the Cilium vxlan interface has any conflicts with other
		// interfaces that makes it unable to brought up, Cilium fails to apply configuration
		// changes. We can remove this block when the following issue gets settled:
		// https://github.com/cilium/cilium/issues/38581
		if vxlanDevice.Name == ciliumVXLANDeviceName && devicePort != port {
			return fmt.Errorf("interface %s uses a different destination port (%d) than the provided config (%d). Please consider adjusting the cluster configuration or removing that device manually", vxlanDevice.Name, devicePort, port)
		}
	}

	return nil
}

// ApplyNetwork will deploy Cilium when network.Enabled is true.
// ApplyNetwork will remove Cilium when network.Enabled is false.
// ApplyNetwork requires that bpf and cgroups2 are already mounted and available when running under strict snap confinement. If they are not, it will fail (since Cilium will not have the required permissions to mount them).
// ApplyNetwork requires that `/sys` is mounted as a shared mount when running under classic snap confinement. This is to ensure that Cilium will be able to automatically mount bpf and cgroups2 on the pods.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyNetwork(ctx context.Context, snap snap.Snap, s state.State, apiserver types.APIServer, network types.Network, annotations types.Annotations) (types.FeatureStatus, error) {
	m := snap.HelmClient()

	if !network.GetEnabled() {
		if _, err := m.Apply(ctx, ChartCilium, helm.StateDeleted, nil); err != nil {
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

	config, err := internalConfig(annotations)
	if err != nil {
		err = fmt.Errorf("failed to parse annotations: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	localhostAddress, err := utils.GetLocalhostAddress()
	if err != nil {
		err = fmt.Errorf("failed to determine localhost address: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		err = fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
	}

	ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	if err != nil {
		err = fmt.Errorf("invalid kube-proxy --cluster-cidr value: %w", err)
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
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

	if err := checkAndSanitizeCiliumVXLAN(config.tunnelPort); err != nil {
		return types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
		}, err
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
			"confPath":     "/etc/cni/net.d",
			"binPath":      "/opt/cni/bin",
			"exclusive":    config.cniExclusive,
			"chainingMode": "portmap",
		},
		"sctp": map[string]any{
			"enabled": config.sctpEnabled,
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
		"envoy": map[string]any{
			"enabled": false, // 1.16+ installs envoy as a standalone daemonset by default if not explicitly disabled
		},
		// https://docs.cilium.io/en/v1.15/network/kubernetes/kubeproxy-free/#kube-proxy-hybrid-modes
		"nodePort":                 ciliumNodePortValues,
		"disableEnvoyVersionCheck": true,
		// socketLB requires an endpoint to the apiserver that's not managed by the kube-proxy
		// so we point to the localhost:secureport to talk to either the kube-apiserver or the kube-apiserver-proxy
		"k8sServiceHost": strings.Trim(localhostAddress.String(), "[]"), // Cilium already adds the brackets for ipv6 addresses, so we need to remove them
		"k8sServicePort": apiserver.GetSecurePort(),
		// This flag enables the runtime device detection which is set to true by default in Cilium 1.16+
		"enableRuntimeDeviceDetection": true,
		"sessionAffinity":              true,
		"loadBalancer": map[string]any{
			"protocolDifferentiation": map[string]any{
				"enabled": true,
			},
		},
		"tunnelPort": config.tunnelPort,
	}

	// Revert these values to default in case they were changed in previous versions
	if ipv4CIDR == "" && ipv6CIDR != "" {
		values["routingMode"] = "tunnel"
		values["ipv6NativeRoutingCIDR"] = ""
		values["autoDirectNodeRoutes"] = false
	}

	if config.devices != "" {
		values["devices"] = config.devices
	}

	if snap.Strict() {
		bpfMnt, err := GetMountPath("bpf")
		if err != nil {
			err = fmt.Errorf("failed to get bpf mount path: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
			}, err
		}

		cgrMnt, err := GetMountPath("cgroup2")
		if err != nil {
			err = fmt.Errorf("failed to get cgroup2 mount path: %w", err)
			return types.FeatureStatus{
				Enabled: false,
				Version: CiliumAgentImageTag,
				Message: fmt.Sprintf(NetworkDeployFailedMsgTmpl, err),
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

	if _, err := m.Apply(ctx, ChartCilium, helm.StatePresent, values); err != nil {
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
