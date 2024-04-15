package k8s

import (
	"fmt"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newGetCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		outputFormat string
	}
	cmd := &cobra.Command{
		Use:    "get <feature.key>",
		Short:  "Get cluster configuration",
		Long:   fmt.Sprintf("Show configuration of one of %s.", strings.Join(componentList, ", ")),
		Args:   cmdutil.MaximumNArgs(env, 1),
		PreRun: chainPreRunHooks(hookRequireRoot(env), hookInitializeFormatter(env, &opts.outputFormat)),
		Run: func(cmd *cobra.Command, args []string) {
			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			config, err := client.GetClusterConfig(cmd.Context(), api.GetClusterConfigRequest{})
			if err != nil {
				cmd.PrintErrf("Error: Failed to get the current cluster configuration.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			// ignore MetricsServer config
			config.MetricsServer = api.MetricsServerConfig{}

			var key string
			if len(args) == 1 {
				key = args[0]
			}

			var output any
			switch key {
			case "":
				output = config
			case "network":
				output = config.Network
			case "dns":
				output = config.DNS
			case "gateway":
				output = config.Gateway
			case "ingress":
				output = config.Ingress
			case "local-storage":
				output = config.LocalStorage
			case "load-balancer":
				output = config.LoadBalancer
			case "network.enabled":
				output = config.Network.GetEnabled()
			case "dns.enabled":
				output = config.DNS.GetEnabled()
			case "dns.upstream-nameservers":
				output = config.DNS.GetUpstreamNameservers()
			case "dns.cluster-domain":
				output = config.DNS.GetClusterDomain()
			case "dns.service-ip":
				output = config.DNS.GetServiceIP()
			case "gateway.enabled":
				output = config.Gateway.GetEnabled()
			case "ingress.enabled":
				output = config.Ingress.GetEnabled()
			case "ingress.default-tls-secret":
				output = config.Ingress.GetDefaultTLSSecret()
			case "ingress.enable-proxy-protocol":
				output = config.Ingress.GetEnableProxyProtocol()
			case "local-storage.enabled":
				output = config.LocalStorage.GetEnabled()
			case "local-storage.local-path":
				output = config.LocalStorage.GetLocalPath()
			case "local-storage.reclaim-policy":
				output = config.LocalStorage.GetReclaimPolicy()
			case "local-storage.default":
				output = config.LocalStorage.GetDefault()
			case "load-balancer.enabled":
				output = config.LoadBalancer.GetEnabled()
			case "load-balancer.cidrs":
				output = config.LoadBalancer.GetCIDRs()
			case "load-balancer.l2-mode":
				output = config.LoadBalancer.GetL2Mode()
			case "load-balancer.l2-interfaces":
				output = config.LoadBalancer.GetL2Interfaces()
			case "load-balancer.bgp-mode":
				output = config.LoadBalancer.GetBGPMode()
			case "load-balancer.bgp-local-asn":
				output = config.LoadBalancer.GetBGPLocalASN()
			case "load-balancer.bgp-peer-address":
				output = config.LoadBalancer.GetBGPPeerAddress()
			case "load-balancer.bgp-peer-port":
				output = config.LoadBalancer.GetBGPPeerPort()
			case "load-balancer.bgp-peer-asn":
				output = config.LoadBalancer.GetBGPPeerASN()
			default:
				cmd.PrintErrf("Error: Unknown config key %q.\n", key)
				env.Exit(1)
				return
			}

			outputFormatter.Print(output)
		},
	}
	cmd.Flags().StringVar(&opts.outputFormat, "output-format", "plain", "set the output format to one of plain, json or yaml")

	return cmd
}
