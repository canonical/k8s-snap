package k8s

import (
	"fmt"
	"strconv"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

func newSetCmd() *cobra.Command {
	setCmd := &cobra.Command{
		Use:               "set <functionality.key=value>...",
		Short:             "Set functionality configuration",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		Args:              cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			config := api.UserFacingClusterConfig{}

			for _, arg := range args {
				configOption, value := splitConfigAndValue(arg)
				switch configOption {
				case "network.enabled":
					if config.Network == nil {
						config.Network = &api.NetworkConfig{}
					}
					if err := setBoolArgument(value, &config.Network.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for network.enabled: %w", err)
					}
				case "dns.enabled":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					if err := setBoolArgument(value, &config.DNS.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for dns.enabled: %w", err)
					}
				case "dns.upstream-nameservers":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					if err := setStringSliceArgument(value, &config.DNS.UpstreamNameservers); err != nil {
						return fmt.Errorf("invalid string slice value for dns.upstream-nameservers: %w", err)
					}
				case "dns.cluster-domain":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					if err := setStringArgument(value, &config.DNS.ClusterDomain); err != nil {
						return fmt.Errorf("invalid string value for dns.cluster-domain: %w", err)
					}
				case "dns.service-ip":
					if config.DNS == nil {
						config.DNS = &api.DNSConfig{}
					}
					if err := setStringArgument(value, &config.DNS.ServiceIP); err != nil {
						return fmt.Errorf("invalid string value for dns.service-ip: %w", err)
					}
				case "gateway.enabled":
					if config.Gateway == nil {
						config.Gateway = &api.GatewayConfig{}
					}
					if err := setBoolArgument(value, &config.Gateway.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for gateway.enabled: %w", err)
					}
				case "ingress.enabled":
					if config.Ingress == nil {
						config.Ingress = &api.IngressConfig{}
					}
					if err := setBoolArgument(value, &config.Ingress.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for ingress.enabled: %w", err)
					}
				case "ingress.default-tls-secret":
					if config.Ingress == nil {
						config.Ingress = &api.IngressConfig{}
					}
					if err := setStringArgument(value, &config.Ingress.DefaultTLSSecret); err != nil {
						return fmt.Errorf("invalid string value for ingress.default-tls-secret: %w", err)
					}
				case "ingress.enable-proxy-protocol":
					if config.Ingress == nil {
						config.Ingress = &api.IngressConfig{}
					}
					if err := setBoolArgument(value, &config.Ingress.EnableProxyProtocol); err != nil {
						return fmt.Errorf("invalid boolean value for ingress.enable-proxy-protocol: %w", err)
					}
				case "local-storage.enabled":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					if err := setBoolArgument(value, &config.LocalStorage.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for local-storage.enabled: %w", err)
					}
				case "local-storage.local-path":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					if err := setStringArgument(value, &config.LocalStorage.LocalPath); err != nil {
						return fmt.Errorf("invalid string value for local-storage.local-path: %w", err)
					}
				case "local-storage.reclaim-policy":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					if err := setStringArgument(value, &config.LocalStorage.ReclaimPolicy); err != nil {
						return fmt.Errorf("invalid string value for local-storage.reclaim-policy: %w", err)
					}
				case "local-storage.set-default":
					if config.LocalStorage == nil {
						config.LocalStorage = &api.LocalStorageConfig{}
					}
					if err := setBoolArgument(value, &config.LocalStorage.SetDefault); err != nil {
						return fmt.Errorf("invalid boolean value for local-storage.set-default: %w", err)
					}
				case "load-balancer.enabled":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setBoolArgument(value, &config.LoadBalancer.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for load-balancer.enabled: %w", err)
					}
				case "load-balancer.cidrs":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setStringSliceArgument(value, &config.LoadBalancer.CIDRs); err != nil {
						return fmt.Errorf("invalid string slice value for load-balancer.cidrs: %w", err)
					}
				case "load-balancer.l2-mode":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setBoolArgument(value, &config.LoadBalancer.L2Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for load-balancer.l2-mode: %w", err)
					}
				case "load-balancer.l2-interfaces":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setStringSliceArgument(value, &config.LoadBalancer.L2Interfaces); err != nil {
						return fmt.Errorf("invalid string slice value for load-balancer.l2-interfaces: %w", err)
					}
				case "load-balancer.bgp-mode":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setBoolArgument(value, &config.LoadBalancer.BGPEnabled); err != nil {
						return fmt.Errorf("invalid boolean value for load-balancer.bgp-mode: %w", err)
					}
				case "load-balancer.bgp-local-asn":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setIntArgument(value, &config.LoadBalancer.BGPLocalASN); err != nil {
						return fmt.Errorf("invalid integer value for load-balancer.bgp-local-asn: %w", err)
					}
				case "load-balancer.bgp-peer-address":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setStringArgument(value, &config.LoadBalancer.BGPPeerAddress); err != nil {
						return fmt.Errorf("invalid string value for load-balancer.bgp-peer-address: %w", err)
					}
				case "load-balancer.bgp-peer-port":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setIntArgument(value, &config.LoadBalancer.BGPPeerPort); err != nil {
						return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-port: %w", err)
					}
				case "load-balancer.bgp-peer-asn":
					if config.LoadBalancer == nil {
						config.LoadBalancer = &api.LoadBalancerConfig{}
					}
					if err := setIntArgument(value, &config.LoadBalancer.BGPPeerASN); err != nil {
						return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-asn: %w", err)
					}
				case "metrics-server.enabled":
					if config.MetricsServer == nil {
						config.MetricsServer = &api.MetricsServerConfig{}
					}
					if err := setBoolArgument(value, &config.MetricsServer.Enabled); err != nil {
						return fmt.Errorf("invalid boolean value for metrics-server.enabled: %w", err)
					}
				default:
					return fmt.Errorf("invalid argument: %s", configOption)
				}
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

func splitConfigAndValue(arg string) (string, string) {
	splitArg := strings.Split(arg, "=")
	return splitArg[0], splitArg[1]
}

func setBoolArgument(val string, target **bool) error {
	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("invalid boolean value: %w", err)
	}
	*target = &parsed
	return nil
}

func setIntArgument(val string, target *int) error {
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid int value: %w", err)
	}
	*target = parsed
	return nil
}

func setStringSliceArgument(val string, target *[]string) error {
	parsed := strings.FieldsFunc(val, func(r rune) bool {
		return r == ','
	})
	*target = parsed
	return nil
}

func setStringArgument(val string, target *string) error {
	*target = val
	return nil
}
