package k8s

import (
	"fmt"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/spf13/cobra"
)

var enableLoadBalancerCmdOpts struct {
	CIDRs          []string
	L2Enabled      bool
	L2Interfaces   []string
	BGPEnabled     bool
	BGPLocalASN    int
	BGPPeerAddress string
	BGPPeerASN     int
	BGPPeerPort    int
}

func newEnableLoadBalancerCmd() *cobra.Command {
	enableLoadBalancerCmd := &cobra.Command{
		Use:     "loadbalancer",
		Short:   "Enable the LoadBalancer component in the cluster",
		PreRunE: chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			request := api.UpdateLoadBalancerComponentRequest{
				Status: api.ComponentEnable,
				Config: api.LoadBalancerComponentConfig{
					CIDRs:          enableLoadBalancerCmdOpts.CIDRs,
					L2Enabled:      enableLoadBalancerCmdOpts.L2Enabled,
					L2Interfaces:   enableLoadBalancerCmdOpts.L2Interfaces,
					BGPEnabled:     enableLoadBalancerCmdOpts.BGPEnabled,
					BGPPeerAddress: enableLoadBalancerCmdOpts.BGPPeerAddress,
					BGPPeerASN:     enableLoadBalancerCmdOpts.BGPPeerASN,
					BGPPeerPort:    enableLoadBalancerCmdOpts.BGPPeerPort,
				},
			}

			if err := k8sdClient.UpdateLoadBalancerComponent(cmd.Context(), request); err != nil {
				return fmt.Errorf("failed to enable LoadBalancer component: %w", err)
			}

			cmd.Println("Component 'LoadBalancer' enabled")
			return nil
		},
	}

	enableLoadBalancerCmd.Flags().StringSliceVar(&enableLoadBalancerCmdOpts.CIDRs, "cidrs", []string{}, "List of CIDRs that will be used for LoadBalancer IP addresses.")
	enableLoadBalancerCmd.MarkFlagRequired("cidrs")
	enableLoadBalancerCmd.Flags().BoolVar(&enableLoadBalancerCmdOpts.L2Enabled, "l2-mode", true, "If set, L2 mode will be enabled for the LoadBalancer")
	enableLoadBalancerCmd.Flags().StringSliceVar(&enableLoadBalancerCmdOpts.L2Interfaces, "l2-interfaces", []string{}, "List of interface names that will be used to announce services in L2 mode. All interfaces used by default.")
	enableLoadBalancerCmd.Flags().BoolVar(&enableLoadBalancerCmdOpts.BGPEnabled, "bgp-mode", false, "If set, BGP mode will be enabled for the LoadBalancer")
	enableLoadBalancerCmd.Flags().IntVar(&enableLoadBalancerCmdOpts.BGPLocalASN, "bgp-local-asn", 64512, "ASN number to use for the cluster's BGP router.")
	enableLoadBalancerCmd.Flags().StringVar(&enableLoadBalancerCmdOpts.BGPPeerAddress, "bgp-peer-address", "", "Address(with slash notation) of the BGP peer.")
	enableLoadBalancerCmd.Flags().IntVar(&enableLoadBalancerCmdOpts.BGPPeerASN, "bgp-peer-asn", 0, "ASN number of the BGP peer.")
	enableLoadBalancerCmd.Flags().IntVar(&enableLoadBalancerCmdOpts.BGPPeerPort, "bgp-peer-port", 0, "Port number of the BGP peer.")
	enableLoadBalancerCmd.MarkFlagsRequiredTogether("bgp-mode", "bgp-local-asn", "bgp-peer-address", "bgp-peer-asn", "bgp-peer-port")
	return enableLoadBalancerCmd
}
