package k8s

import (
	"context"
	"fmt"
	"strings"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/cmd/k8s/formatter"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	getCmd := &cobra.Command{
		Use:               "get <functionality.key>",
		Short:             "get functionality configuration",
		PersistentPreRunE: chainPreRunHooks(hookSetupClient),
		Args:              cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			field := args[0]

			timeoutCtx, cancel := context.WithTimeout(cmd.Context(), statusCmdOpts.timeout)
			defer cancel()

			clusterConfig, err := k8sdClient.GetClusterConfig(timeoutCtx, api.GetClusterConfigRequest{})
			if err != nil {
				return fmt.Errorf("failed to get cluster config: %w", err)
			}

			if !strings.Contains(field, ".") {
				f, err := formatter.New("yaml", cmd.OutOrStdout())
				if err != nil {
					return fmt.Errorf("failed to create formatter: %w", err)
				}
				switch field {
				case "network":
					return f.Print(clusterConfig.Network)
				case "dns":
					return f.Print(clusterConfig.DNS)
				case "gateway":
					return f.Print(clusterConfig.Gateway)
				case "ingress":
					return f.Print(clusterConfig.Ingress)
				case "local-storage":
					return f.Print(clusterConfig.LocalStorage)
				case "load-balancer":
					return f.Print(clusterConfig.LoadBalancer)
				case "metrics-server":
					return f.Print(clusterConfig.MetricsServer)
				default:
					return fmt.Errorf("invalid argument: %s", field)
				}
			} else {
				f, err := formatter.New("plain", cmd.OutOrStdout())
				if err != nil {
					return fmt.Errorf("failed to create formatter: %w", err)
				}

				switch field {
				case "network.enabled":
					return f.Print(*clusterConfig.Network.Enabled)
				case "dns.enabled":
					return f.Print(*clusterConfig.DNS.Enabled)
				case "dns.upstream-nameservers":
					return f.Print(clusterConfig.DNS.UpstreamNameservers)
				case "dns.cluster-domain":
					return f.Print(clusterConfig.DNS.ClusterDomain)
				case "dns.service-ip":
					return f.Print(clusterConfig.DNS.ServiceIP)
				case "gateway.enabled":
					return f.Print(*clusterConfig.Gateway.Enabled)
				case "ingress.enabled":
					return f.Print(*clusterConfig.Ingress.Enabled)
				case "ingress.default-tls-secret":
					return f.Print(clusterConfig.Ingress.DefaultTLSSecret)
				case "ingress.enable-proxy-protocol":
					return f.Print(*clusterConfig.Ingress.EnableProxyProtocol)
				case "local-storage.enabled":
					return f.Print(*clusterConfig.LocalStorage.Enabled)
				case "local-storage.local-path":
					return f.Print(clusterConfig.LocalStorage.LocalPath)
				case "local-storage.reclaim-policy":
					return f.Print(clusterConfig.LocalStorage.ReclaimPolicy)
				case "local-storage.set-default":
					return f.Print(*clusterConfig.LocalStorage.SetDefault)
				case "load-balancer.enabled":
					return f.Print(*clusterConfig.LoadBalancer.Enabled)
				case "load-balancer.cidrs":
					return f.Print(clusterConfig.LoadBalancer.CIDRs)
				case "load-balancer.l2-mode":
					return f.Print(*clusterConfig.LoadBalancer.L2Enabled)
				case "load-balancer.l2-interfaces":
					return f.Print(clusterConfig.LoadBalancer.L2Interfaces)
				case "load-balancer.bgp-mode":
					return f.Print(*clusterConfig.LoadBalancer.BGPEnabled)
				case "load-balancer.bgp-local-asn":
					return f.Print(clusterConfig.LoadBalancer.BGPLocalASN)
				case "load-balancer.bgp-peer-address":
					return f.Print(clusterConfig.LoadBalancer.BGPPeerAddress)
				case "load-balancer.bgp-peer-port":
					return f.Print(clusterConfig.LoadBalancer.BGPPeerPort)
				case "load-balancer.bgp-peer-asn":
					return f.Print(clusterConfig.LoadBalancer.BGPPeerASN)
				case "metrics-server.enabled":
					return f.Print(*clusterConfig.MetricsServer.Enabled)
				default:
					return fmt.Errorf("invalid argument: %s", field)
				}
			}
		},
	}

	return getCmd
}
