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
	"github.com/canonical/lxd/shared/api"
)

const (
	CiliumDisableFailedMsgTmpl = "Failed to disable Cilium features, the error was: %v"
	CiliumEnableFailedMsgTmpl  = "Failed to enable Cilium features, the error was: %v"

	IngressOptionEnabled                          = "enabled"
	IngressOptionLoadBalancerMode                 = "loadbalancerMode"
	IngressOptionLoadBalancerModeShared           = "shared" // loadbalancerMode: "shared"
	IngressOptionDefaultSecretName                = "defaultSecretName"
	IngressOptionDefaultSecretNamespace           = "defaultSecretNamespace"
	IngressOptionDefaultSecretNamespaceKubeSystem = "kube-system" // defaultSecretNamespace: "kube-system"
	IngressOptionEnableProxyProtocol              = "enableProxyProtocol"
)

// required for unittests.
var (
	GetMountPath            = utils.GetMountPath
	GetMountPropagationType = utils.GetMountPropagationType
)

type AddressGetter interface {
	Address() *api.URL
}

func ApplyCilium(
	ctx context.Context,
	snap snap.Snap,
	addrGetter AddressGetter,
	apiServer types.APIServer,
	network types.Network,
	gateway types.Gateway,
	ingress types.Ingress,
	annotations types.Annotations,
) (map[types.FeatureName]types.FeatureStatus, error) {
	// network
	m := snap.HelmClient()

	if !network.GetEnabled() {
		if _, err := m.Apply(ctx, ChartCilium, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to uninstall cilium chart: %w", err)
			return returnStatuses(
				false,
				false,
				false,
				err,
			), err

		}
		return returnStatuses(
			false,
			false,
			false,
			nil,
		), nil
	}

	config, err := internalConfig(annotations)
	if err != nil {
		err = fmt.Errorf("failed to parse annotations: %w", err)
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
	}

	localhostAddress, err := utils.GetLocalhostAddress()
	if err != nil {
		err = fmt.Errorf("failed to get localhost address: %w", err)
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
	}

	nodeIP := net.ParseIP(addrGetter.Address().Hostname())
	if nodeIP == nil {
		err = fmt.Errorf("failed to parse node IP address %q", addrGetter.Address().Hostname())
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
	}

	defaultCidr, err := utils.FindCIDRForIP(nodeIP)
	if err != nil {
		err = fmt.Errorf("failed to find cidr of default interface: %w", err)
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
	}

	ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	if err != nil {
		err = fmt.Errorf("invalid kube-proxy --cluster-cidr value: %w", err)
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
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
		err = fmt.Errorf("failed to check and sanitize Cilium VXLAN port: %w", err)
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
	}

	bpfValues := map[string]any{}
	if config.vlanBPFBypass != nil {
		bpfValues["vlanBypass"] = config.vlanBPFBypass
	}

	ciliumValues := map[string]any{
		"bpf": bpfValues,
		"image": map[string]any{
			"repository": CiliumAgentImageRepo,
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
				"repository": CiliumOperatorImageRepo,
				"tag":        CiliumOperatorImageTag,
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
		"k8sServicePort": apiServer.GetSecurePort(),
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

	// If we are deploying with IPv6 only, we need to set the routing mode to native
	if ipv4CIDR == "" && ipv6CIDR != "" {
		ciliumValues["routingMode"] = "native"
		ciliumValues["ipv6NativeRoutingCIDR"] = defaultCidr
		ciliumValues["autoDirectNodeRoutes"] = true
	}

	if config.devices != "" {
		ciliumValues["devices"] = config.devices
	}

	if snap.Strict() {
		bpfMnt, err := GetMountPath("bpf")
		if err != nil {
			err = fmt.Errorf("failed to get bpf mount path: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}

		cgrMnt, err := GetMountPath("cgroup2")
		if err != nil {
			err = fmt.Errorf("failed to get cgroup2 mount path: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}

		ciliumValues["bpf"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"root": bpfMnt,
		}
		ciliumValues["cgroup"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"hostRoot": cgrMnt,
		}
	} else {
		pt, err := GetMountPropagationType("/sys")
		if err != nil {
			err = fmt.Errorf("failed to get mount propagation type for /sys: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}
		if pt == utils.MountPropagationPrivate {
			onLXD, err := snap.OnLXD(ctx)
			if err != nil {
				logger := log.FromContext(ctx)
				logger.Error(err, "Failed to check if running on LXD")
			}
			if onLXD {
				err := fmt.Errorf("/sys is not a shared mount on the LXD container, this might be resolved by updating LXD on the host to version 5.0.2 or newer")
				return returnStatuses(
					network.GetEnabled(),
					gateway.GetEnabled(),
					ingress.GetEnabled(),
					err,
				), err
			}

			err = fmt.Errorf("/sys is not a shared mount")
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}
	}

	// gateway
	if gateway.GetEnabled() {
		// Install Gateway API CRDs
		if _, err := m.Apply(ctx, chartGateway, helm.StatePresent, nil); err != nil {
			err = fmt.Errorf("failed to install Gateway API CRDs: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}

		// Apply our GatewayClass named ck-gateway
		if _, err := m.Apply(ctx, chartGatewayClass, helm.StatePresent, nil); err != nil {
			err = fmt.Errorf("failed to install Gateway API GatewayClass: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}

		ciliumValues["gatewayAPI"] = map[string]any{
			"enabled": true,
			"gatewayClass": map[string]any{
				// This needs to be string, not bool, as the helm chart uses a string
				// Due to the values of 'auto', 'true' and 'false'
				"create": "false",
			},
		}
	} else {
		// Delete our GatewayClass named ck-gateway
		if _, err := m.Apply(ctx, chartGatewayClass, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to delete Gateway API GatewayClass: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}

		ciliumValues["gatewayAPI"] = map[string]any{"enabled": false}
	}

	// ingress
	if ingress.GetEnabled() {
		ciliumValues["ingressController"] = map[string]any{
			IngressOptionEnabled:                true,
			IngressOptionLoadBalancerMode:       IngressOptionLoadBalancerModeShared,
			IngressOptionDefaultSecretNamespace: IngressOptionDefaultSecretNamespaceKubeSystem,
			IngressOptionDefaultSecretName:      ingress.GetDefaultTLSSecret(),
			IngressOptionEnableProxyProtocol:    ingress.GetEnableProxyProtocol(),
		}
	} else {
		ciliumValues["ingressController"] = map[string]any{
			IngressOptionEnabled:                false,
			IngressOptionLoadBalancerMode:       "",
			IngressOptionDefaultSecretNamespace: "",
			IngressOptionDefaultSecretName:      "",
			IngressOptionEnableProxyProtocol:    false,
		}
	}

	changed, err := m.Apply(ctx, ChartCilium, helm.StatePresent, ciliumValues)
	if err != nil {
		err = fmt.Errorf("failed to apply cilium chart: %w", err)
		return returnStatuses(
			network.GetEnabled(),
			gateway.GetEnabled(),
			ingress.GetEnabled(),
			err,
		), err
	}

	if !gateway.GetEnabled() {
		// Remove Gateway CRDs if the Gateway feature is disabled.
		// This is done after the Cilium update as cilium requires the CRDs to be present for cleanups.
		if _, err := m.Apply(ctx, chartGateway, helm.StateDeleted, nil); err != nil {
			err = fmt.Errorf("failed to delete Gateway API CRDs: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}
	}

	if changed {
		if err := rolloutRestartCilium(ctx, snap, 3); err != nil {
			err = fmt.Errorf("failed to rollout restart cilium: %w", err)
			return returnStatuses(
				network.GetEnabled(),
				gateway.GetEnabled(),
				ingress.GetEnabled(),
				err,
			), err
		}
	}

	return returnStatuses(
		true,
		gateway.GetEnabled(),
		ingress.GetEnabled(),
		nil,
	), nil
}

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

func returnStatuses(networkEnabled, gatewayEnabled, ingressEnabled bool, err error) map[types.FeatureName]types.FeatureStatus {
	ss := map[types.FeatureName]types.FeatureStatus{}

	// network
	if networkEnabled {
		enabledMsg := EnabledMsg
		enabled := true
		if err != nil {
			enabledMsg = fmt.Sprintf(CiliumEnableFailedMsgTmpl, err)
			enabled = false
		}
		ss[types.FeatureName("network")] = types.FeatureStatus{
			Enabled: enabled,
			Version: CiliumAgentImageTag,
			Message: enabledMsg,
		}
	} else {
		disabledMsg := DisabledMsg
		if err != nil {
			disabledMsg = fmt.Sprintf(CiliumDisableFailedMsgTmpl, err)
		}
		ss[types.FeatureName("network")] = types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: disabledMsg,
		}
	}

	// gateway
	if gatewayEnabled {
		enabledMsg := EnabledMsg
		enabled := true
		if err != nil {
			enabledMsg = fmt.Sprintf(CiliumEnableFailedMsgTmpl, err)
			enabled = false
		}
		ss[types.FeatureName("gateway")] = types.FeatureStatus{
			Enabled: enabled,
			Version: CiliumAgentImageTag,
			Message: enabledMsg,
		}
	} else {
		disabledMsg := DisabledMsg
		if err != nil {
			disabledMsg = fmt.Sprintf(CiliumDisableFailedMsgTmpl, err)
		}
		ss[types.FeatureName("gateway")] = types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: disabledMsg,
		}
	}

	// ingress
	if ingressEnabled {
		enabledMsg := EnabledMsg
		enabled := true
		if err != nil {
			enabledMsg = fmt.Sprintf(CiliumEnableFailedMsgTmpl, err)
			enabled = false
		}
		ss[types.FeatureName("ingress")] = types.FeatureStatus{
			Enabled: enabled,
			Version: CiliumAgentImageTag,
			Message: enabledMsg,
		}
	} else {
		disabledMsg := DisabledMsg
		if err != nil {
			disabledMsg = fmt.Sprintf(CiliumDisableFailedMsgTmpl, err)
		}
		ss[types.FeatureName("ingress")] = types.FeatureStatus{
			Enabled: false,
			Version: CiliumAgentImageTag,
			Message: disabledMsg,
		}
	}

	return ss
}
