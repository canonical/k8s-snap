package k8s

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	apiv1 "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/spf13/cobra"
)

type SetResult struct {
	ClusterConfig apiv1.UserFacingClusterConfig `json:"cluster-config" yaml:"cluster-config"`
}

func (s SetResult) String() string {
	return "Configuration updated."
}

func newSetCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	return &cobra.Command{
		Use:    "set <functionality.key=value> ...",
		Short:  "Set cluster configuration",
		Long:   fmt.Sprintf("Configure one of %s.\nUse `k8s get` to explore configuration options.", strings.Join(componentList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			config := apiv1.UserFacingClusterConfig{}

			for _, arg := range args {
				if err := updateConfig(&config, arg); err != nil {
					cmd.PrintErrf("ERROR: Invalid option %q.\n\nThe error was: %v\n", arg, err)
				}
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			// Fetching current config to check where an already enabled functionality is updated.
			currentConfig, err := client.GetClusterConfig(cmd.Context(), apiv1.GetClusterConfigRequest{})
			if err != nil {
				cmd.PrintErrf("ERROR: Failed to retrieve the current cluster configuration.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if vals.OptionalBool(currentConfig.Network.Enabled, false) && config.Network != nil && config.Network.Enabled == nil {
				cmd.PrintErrln("network configuration will be updated")
			}
			if vals.OptionalBool(currentConfig.DNS.Enabled, false) && config.DNS != nil && config.DNS.Enabled == nil {
				cmd.PrintErrln("dns configuration will be updated")
			}
			if vals.OptionalBool(currentConfig.Gateway.Enabled, false) && config.Gateway != nil && config.Gateway.Enabled == nil {
				cmd.PrintErrln("gateway configuration will be updated")
			}
			if vals.OptionalBool(currentConfig.Ingress.Enabled, false) && config.Ingress != nil && config.Ingress.Enabled == nil {
				cmd.PrintErrln("ingress configuration will be updated")
			}
			if vals.OptionalBool(currentConfig.LocalStorage.Enabled, false) && config.LocalStorage != nil && config.LocalStorage.Enabled == nil {
				cmd.PrintErrln("local-storage configuration will be updated")
			}
			if vals.OptionalBool(currentConfig.LoadBalancer.Enabled, false) && config.LoadBalancer != nil && config.LoadBalancer.Enabled == nil {
				cmd.PrintErrln("load-balancer configuration will be updated")
			}
			if vals.OptionalBool(currentConfig.MetricsServer.Enabled, false) && config.MetricsServer != nil && config.MetricsServer.Enabled == nil {
				cmd.PrintErrln("metrics-server configuration will be updated")
			}

			request := apiv1.UpdateClusterConfigRequest{
				Config: config,
			}

			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("ERROR: Failed to apply requested cluster configuration changes.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if err := cmdutil.FormatterFromContext(cmd.Context()).Print(SetResult{ClusterConfig: config}); err != nil {
				cmd.PrintErrf("WARNING: Failed to print the cluster configuration result.\n\nThe error was: %v\n", err)
			}
		},
	}
}

func updateConfig(config *apiv1.UserFacingClusterConfig, arg string) error {
	parts := strings.SplitN(arg, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("option not in <key>=<value> format")
	}
	key := parts[0]
	value := parts[1]

	switch key {
	case "network.enabled":
		if config.Network == nil {
			config.Network = &apiv1.NetworkConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for network.enabled: %w", err)
		}
		config.Network.Enabled = &v
	case "dns.enabled":
		if config.DNS == nil {
			config.DNS = &apiv1.DNSConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for dns.enabled: %w", err)
		}
		config.DNS.Enabled = &v
	case "dns.upstream-nameservers":
		if config.DNS == nil {
			config.DNS = &apiv1.DNSConfig{}
		}
		config.DNS.UpstreamNameservers = strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
	case "dns.cluster-domain":
		if config.DNS == nil {
			config.DNS = &apiv1.DNSConfig{}
		}
		config.DNS.ClusterDomain = value
	case "dns.service-ip":
		if config.DNS == nil {
			config.DNS = &apiv1.DNSConfig{}
		}
		config.DNS.ServiceIP = value
	case "gateway.enabled":
		if config.Gateway == nil {
			config.Gateway = &apiv1.GatewayConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for gateway.enabled: %w", err)
		}
		config.Gateway.Enabled = &v
	case "ingress.enabled":
		if config.Ingress == nil {
			config.Ingress = &apiv1.IngressConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for ingress.enabled: %w", err)
		}
		config.Ingress.Enabled = &v
	case "ingress.default-tls-secret":
		if config.Ingress == nil {
			config.Ingress = &apiv1.IngressConfig{}
		}
		config.Ingress.DefaultTLSSecret = value
	case "ingress.enable-proxy-protocol":
		if config.Ingress == nil {
			config.Ingress = &apiv1.IngressConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for ingress.enable-proxy-protocol: %w", err)
		}
		config.Ingress.EnableProxyProtocol = &v
	case "local-storage.enabled":
		if config.LocalStorage == nil {
			config.LocalStorage = &apiv1.LocalStorageConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for local-storage.enabled: %w", err)
		}
		config.LocalStorage.Enabled = &v
	case "local-storage.local-path":
		if config.LocalStorage == nil {
			config.LocalStorage = &apiv1.LocalStorageConfig{}
		}
		config.LocalStorage.LocalPath = value
	case "local-storage.reclaim-policy":
		if config.LocalStorage == nil {
			config.LocalStorage = &apiv1.LocalStorageConfig{}
		}
		config.LocalStorage.ReclaimPolicy = value
	case "local-storage.set-default":
		if config.LocalStorage == nil {
			config.LocalStorage = &apiv1.LocalStorageConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for local-storage.set-default: %w", err)
		}
		config.LocalStorage.SetDefault = &v
	case "load-balancer.enabled":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for load-balancer.enabled: %w", err)
		}
		config.LoadBalancer.Enabled = &v
	case "load-balancer.cidrs":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		config.LoadBalancer.CIDRs = strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
	case "load-balancer.l2-mode":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for load-balancer.l2-mode: %w", err)
		}
		config.LoadBalancer.L2Enabled = &v
	case "load-balancer.l2-interfaces":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		config.LoadBalancer.L2Interfaces = strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
	case "load-balancer.bgp-mode":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for load-balancer.bgp-mode: %w", err)
		}
		config.LoadBalancer.BGPEnabled = &v
	case "load-balancer.bgp-local-asn":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for load-balancer.bgp-local-asn: %w", err)
		}
		config.LoadBalancer.BGPLocalASN = v
	case "load-balancer.bgp-peer-address":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		config.LoadBalancer.BGPPeerAddress = value
	case "load-balancer.bgp-peer-port":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-port: %w", err)
		}
		config.LoadBalancer.BGPPeerPort = v
	case "load-balancer.bgp-peer-asn":
		if config.LoadBalancer == nil {
			config.LoadBalancer = &apiv1.LoadBalancerConfig{}
		}
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-asn: %w", err)
		}
		config.LoadBalancer.BGPPeerASN = v
	case "metrics-server.enabled":
		if config.MetricsServer == nil {
			config.MetricsServer = &apiv1.MetricsServerConfig{}
		}
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for metrics-server.enabled: %w", err)
		}
		config.MetricsServer.Enabled = &v
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}
