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
	var opts struct {
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "set <feature.key=value> ...",
		Short:  "Set cluster configuration",
		Long:   fmt.Sprintf("Configure one of %s.\nUse `k8s get` to explore configuration options.", strings.Join(componentList, ", ")),
		Args:   cmdutil.MinimumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			config := apiv1.UserFacingClusterConfig{}

			for _, arg := range args {
				if err := updateConfig(&config, arg); err != nil {
					cmd.PrintErrf("Error: Invalid option %q.\n\nThe error was: %v\n", arg, err)
					env.Exit(1)
				}
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			request := apiv1.UpdateClusterConfigRequest{
				Config: config,
			}

			if err := client.UpdateClusterConfig(cmd.Context(), request); err != nil {
				cmd.PrintErrf("Error: Failed to apply requested cluster configuration changes.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			outputFormatter.Print(SetResult{ClusterConfig: config})
		},
	}

	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
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
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for network.enabled: %w", err)
		}
		config.Network.Enabled = &v
	case "dns.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for dns.enabled: %w", err)
		}
		config.DNS.Enabled = &v
	case "dns.upstream-nameservers":
		config.DNS.UpstreamNameservers = vals.Pointer(strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' }))
	case "dns.cluster-domain":
		config.DNS.ClusterDomain = vals.Pointer(value)
	case "dns.service-ip":
		config.DNS.ServiceIP = vals.Pointer(value)
	case "gateway.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for gateway.enabled: %w", err)
		}
		config.Gateway.Enabled = &v
	case "ingress.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for ingress.enabled: %w", err)
		}
		config.Ingress.Enabled = &v
	case "ingress.default-tls-secret":
		config.Ingress.DefaultTLSSecret = vals.Pointer(value)
	case "ingress.enable-proxy-protocol":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for ingress.enable-proxy-protocol: %w", err)
		}
		config.Ingress.EnableProxyProtocol = &v
	case "local-storage.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for local-storage.enabled: %w", err)
		}
		config.LocalStorage.Enabled = &v
	case "local-storage.local-path":
		config.LocalStorage.LocalPath = vals.Pointer(value)
	case "local-storage.reclaim-policy":
		config.LocalStorage.ReclaimPolicy = vals.Pointer(value)
	case "local-storage.default":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for local-storage.default: %w", err)
		}
		config.LocalStorage.Default = &v
	case "load-balancer.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for load-balancer.enabled: %w", err)
		}
		config.LoadBalancer.Enabled = &v
	case "load-balancer.cidrs":
		config.LoadBalancer.CIDRs = vals.Pointer(strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' }))
	case "load-balancer.l2-mode":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for load-balancer.l2-mode: %w", err)
		}
		config.LoadBalancer.L2Mode = &v
	case "load-balancer.l2-interfaces":
		config.LoadBalancer.L2Interfaces = vals.Pointer(strings.FieldsFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == ',' }))
	case "load-balancer.bgp-mode":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value for load-balancer.bgp-mode: %w", err)
		}
		config.LoadBalancer.BGPMode = &v
	case "load-balancer.bgp-local-asn":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for load-balancer.bgp-local-asn: %w", err)
		}
		config.LoadBalancer.BGPLocalASN = &v
	case "load-balancer.bgp-peer-address":
		config.LoadBalancer.BGPPeerAddress = vals.Pointer(value)
	case "load-balancer.bgp-peer-port":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-port: %w", err)
		}
		config.LoadBalancer.BGPPeerPort = &v
	case "load-balancer.bgp-peer-asn":
		v, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value for load-balancer.bgp-peer-asn: %w", err)
		}
		config.LoadBalancer.BGPPeerASN = &v
	case "metrics-server.enabled":
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
