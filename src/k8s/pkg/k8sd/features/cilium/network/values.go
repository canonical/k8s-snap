package network

import (
	"context"
	"fmt"
	"net"
	"strings"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/microcluster/v2/state"
)

type Values map[string]any

func (v Values) applyDefaultValues() error {
	values := map[string]any{
		"image": map[string]any{
			"useDigest": false,
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
				"useDigest": false,
			},
		},
		"envoy": map[string]any{
			"enabled": false, // 1.16+ installs envoy as a standalone daemonset by default if not explicitly disabled
		},
		// https://docs.cilium.io/en/v1.15/network/kubernetes/kubeproxy-free/#kube-proxy-hybrid-modes
		"nodePort": map[string]any{
			"enabled":           true,
			"enableHealthCheck": false,
		},
		"disableEnvoyVersionCheck": true,
		// This flag enables the runtime device detection which is set to true by default in Cilium 1.16+
		"enableRuntimeDeviceDetection": true,
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride, mergo.WithTypeCheck); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyImageOverrides(manifest types.FeatureManifest) error {
	ciliumAgentImage := manifest.GetImage(CiliumAgentImageName)
	ciliumOperatorImage := manifest.GetImage(CiliumOperatorImageName)

	values := map[string]any{
		"image": map[string]any{
			"repository": ciliumAgentImage.GetURI(),
			"tag":        ciliumAgentImage.Tag,
		},

		"operator": map[string]any{
			"image": map[string]any{
				"repository": ciliumOperatorImage.GetURI(),
				"tag":        ciliumOperatorImage.Tag,
			},
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge image overrides: %w", err)
	}

	return nil
}

func (v Values) applyClusterConfiguration(ctx context.Context, s state.State, apiserver types.APIServer, network types.Network) error {
	c, err := s.Leader()
	if err != nil {
		return fmt.Errorf("failed to get leader client: %w", err)
	}

	clusterMembers, err := c.GetClusterMembers(ctx)
	if err != nil {
		return fmt.Errorf("failed to get cluster members: %w", err)
	}

	localhostAddress, err := utils.DetermineLocalhostAddress(clusterMembers)
	if err != nil {
		return fmt.Errorf("failed to determine localhost address: %w", err)
	}

	nodeIP := net.ParseIP(s.Address().Hostname())
	if nodeIP == nil {
		return fmt.Errorf("failed to parse node IP address %q", s.Address().Hostname())
	}

	defaultCidr, err := utils.FindCIDRForIP(nodeIP)
	if err != nil {
		return fmt.Errorf("failed to find cidr of default interface: %w", err)
	}

	ipv4CIDR, ipv6CIDR, err := utils.SplitCIDRStrings(network.GetPodCIDR())
	if err != nil {
		return fmt.Errorf("invalid kube-proxy --cluster-cidr value: %w", err)
	}

	values := map[string]any{
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
		// socketLB requires an endpoint to the apiserver that's not managed by the kube-proxy
		// so we point to the localhost:secureport to talk to either the kube-apiserver or the kube-apiserver-proxy
		"k8sServiceHost": strings.Trim(localhostAddress, "[]"), // Cilium already adds the brackets for ipv6 addresses, so we need to remove them
		"k8sServicePort": apiserver.GetSecurePort(),
	}

	// If we are deploying with IPv6 only, we need to set the routing mode to native
	if ipv4CIDR == "" && ipv6CIDR != "" {
		values["routingMode"] = "native"
		values["ipv6NativeRoutingCIDR"] = defaultCidr
		values["autoDirectNodeRoutes"] = true
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge default values: %w", err)
	}

	return nil
}

func (v Values) ApplyStrictOverrides() error {
	bpfMnt, err := GetMountPath("bpf")
	if err != nil {
		return fmt.Errorf("failed to get bpf mount path: %w", err)
	}

	cgrMnt, err := GetMountPath("cgroup2")
	if err != nil {
		return fmt.Errorf("failed to get cgroup2 mount path: %w", err)
	}

	values := map[string]any{
		"bpf": map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"root": bpfMnt,
		},
		"cgroup": map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"hostRoot": cgrMnt,
		},
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge strict overrides values: %w", err)
	}

	return nil
}

func (v Values) ApplyAnnotations(annotations types.Annotations) error {
	config, err := internalConfig(annotations)
	if err != nil {
		return fmt.Errorf("failed to parse annotations: %w", err)
	}

	ciliumNodePortValues := map[string]any{}

	if config.directRoutingDevice != "" {
		ciliumNodePortValues["directRoutingDevice"] = config.directRoutingDevice
	}

	bpfValues := map[string]any{}
	if config.vlanBPFBypass != nil {
		bpfValues["vlanBypass"] = config.vlanBPFBypass
	}

	values := map[string]any{
		"bpf": bpfValues,

		"cni": map[string]any{
			"exclusive": config.cniExclusive,
		},
		"sctp": map[string]any{
			"enabled": config.sctpEnabled,
		},

		"nodePort": ciliumNodePortValues,
	}

	if config.devices != "" {
		values["devices"] = config.devices
	}

	if err := mergo.Merge(&v, Values(values), mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge strict overrides values: %w", err)
	}

	return nil
}
