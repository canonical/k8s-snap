package k8s

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/utils/vals"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:     "set <functionality.key=value> ...",
		Short:   "Set cluster configuration",
		Long:    fmt.Sprintf("Configure one of %s.\nUse `k8s get` to explore configuration options.", strings.Join(componentList, ", ")),
		PreRunE: chainPreRunHooks(hookSetupClient),
		Args:    cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			config := api.UserFacingClusterConfig{}

			for _, arg := range args {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("option %q not in <key>=<value> format", arg)
				}
				key := parts[0]
				value := parts[1]

				switch key {
				case "network.enabled":
					if config.Network == nil {
						config.Network = &api.NetworkConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for network.enabled: %w", err)
					}
					config.Network.Enabled = &v
				case "dns.enabled":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for dns.enabled: %w", err)
					}
					config.DNS.Enabled = &v
				case "dns.upstream-nameservers":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					config.DNS.UpstreamNameservers = strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
				case "dns.cluster-domain":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					config.DNS.ClusterDomain = value
				case "dns.service-ip":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					config.DNS.ServiceIP = value
				case "gateway.enabled":
					if config.Gateway == nil {
						config.Gateway = &api.GatewayConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for gateway.enabled: %w", err)
					}
					config.Gateway.Enabled = &v
				case "ingress.enabled":
					if config.Ingress == nil {
						config.Ingress = &api.IngressConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for ingress.enabled: %w", err)
					}
					config.Ingress.Enabled = &v
				case "ingress.default-tls-secret":
					if config.Ingress == nil {
						config.Ingress = &api.IngressConfig{}
					}
					config.Ingress.DefaultTLSSecret = value
				case "ingress.enable-proxy-protocol":
					if config.Ingress == nil {
						config.Ingress = &api.IngressConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for ingress.enable-proxy-protocol: %w", err)
					}
					config.Ingress.EnableProxyProtocol = &v
				case "local-storage.enabled":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for local-storage.enabled: %w", err)
					}
					config.LocalStorage.Enabled = &v
				case "local-storage.local-path":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					config.LocalStorage.LocalPath = value
				case "local-storage.reclaim-policy":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					config.LocalStorage.ReclaimPolicy = value
				case "local-storage.set-default":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for local-storage.set-default: %w", err)
					}
					config.LocalStorage.SetDefault = &v
				case "load-balancer.enabled":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for load-balancer.enabled: %w", err)
					}
					config.LoadBalancer.Enabled = &v
				case "load-balancer.cidrs":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					config.LoadBalancer.CIDRs = strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
				case "load-balancer.l2-mode":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for load-balancer.l2-mode: %w", err)
					}
					config.LoadBalancer.L2Enabled = &v
				case "load-balancer.l2-interfaces":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					config.LoadBalancer.L2Interfaces = strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
				case "load-balancer.bgp-mode":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for load-balancer.bgp-mode: %w", err)
					}
					config.LoadBalancer.BGPEnabled = &v
				case "load-balancer.bgp-local-asn":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					v, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("invalid integer value for load-balancer.bgp-local-asn: %w", err)
					}
					config.LoadBalancer.BGPLocalASN = v
				case "load-balancer.bgp-peer-address":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					config.LoadBalancer.BGPPeerAddress = value
				case "load-balancer.bgp-peer-port":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					v, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-port: %w", err)
					}
					config.LoadBalancer.BGPPeerPort = v
				case "load-balancer.bgp-peer-asn":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					v, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-asn: %w", err)
					}
					config.LoadBalancer.BGPPeerASN = v
				case "metrics-server.enabled":
					if config.MetricsServer == nil {
						config.MetricsServer = &api.MetricsServerConfig{}
					}
					v, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for metrics-server.enabled: %w", err)
					}
					config.MetricsServer.Enabled = &v
				default:
					return fmt.Errorf("invalid config key: %s", key)
				}
			}

			// Fetching current config to check where an already enabled functionality is updated.
			currentConfig, err := k8sdClient.GetClusterConfig(cmd.Context(), api.GetClusterConfigRequest{})
			if err != nil {
				return fmt.Errorf("failed to get current cluster config: %w", err)
			}

			if vals.OptionalBool(currentConfig.Network.Enabled, false) && config.Network != nil && config.Network.Enabled == nil {
				fmt.Println("Reapplying configuration for network")
			}
			if vals.OptionalBool(currentConfig.DNS.Enabled, false) && config.DNS != nil && config.DNS.Enabled == nil {
				fmt.Println("Reapplying configuration for dns")
			}
			if vals.OptionalBool(currentConfig.Gateway.Enabled, false) && config.Gateway != nil && config.Gateway.Enabled == nil {
				fmt.Println("Reapplying configuration for gateway")
			}
			if vals.OptionalBool(currentConfig.Ingress.Enabled, false) && config.Ingress != nil && config.Ingress.Enabled == nil {
				fmt.Println("Reapplying configuration for ingress")
			}
			if vals.OptionalBool(currentConfig.LocalStorage.Enabled, false) && config.LocalStorage != nil && config.LocalStorage.Enabled == nil {
				fmt.Println("Reapplying configuration for local-storage")
			}
			if vals.OptionalBool(currentConfig.LoadBalancer.Enabled, false) && config.LoadBalancer != nil && config.LoadBalancer.Enabled == nil {
				fmt.Println("Reapplying configuration for load-balancer")
			}
			if vals.OptionalBool(currentConfig.MetricsServer.Enabled, false) && config.MetricsServer != nil && config.MetricsServer.Enabled == nil {
				fmt.Println("Reapplying configuration for metrics-server")
			}

			request := api.UpdateClusterConfigRequest{
				Config: config,
			}

			if err := k8sdClient.UpdateClusterConfig(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to update cluster configuration: %w", err)
			}
			return nil
		},
	}

	return setCmd
}
