package k8s

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var (
	componentList = []string{"network", "dns", "gateway", "ingress", "rbac", "storage", "loadbalancer"}
)

func newEnableCmd() *cobra.Command {
	enableCmd := &cobra.Command{
		Use:   "enable <component>",
		Short: "Enable a specific component in the cluster",
		Long:  fmt.Sprintf("Enable one of the specific components: %s.", strings.Join(componentList, ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("Too many arguments. Please, only provide the name of the component that should be enabled.")
			}
			if len(args) < 1 {
				return fmt.Errorf("Not enough arguments. Please, provide the name of the component that should be enabled.")
			}
			if !slices.Contains(componentList, args[0]) {
				return fmt.Errorf("Unknown component %q. Needs to be one of: %s", args[0], strings.Join(componentList, ", "))
			}
			return nil
		},
	}
	enableCmd.AddCommand(newEnableDNSCmd())
	enableCmd.AddCommand(newEnableNetworkCmd())
	enableCmd.AddCommand(newEnableStorageCmd())
	enableCmd.AddCommand(newEnableIngressCmd())
	enableCmd.AddCommand(newEnableGatewayCmd())
	enableCmd.AddCommand(newEnableLoadBalancerCmd())
	return enableCmd
}
