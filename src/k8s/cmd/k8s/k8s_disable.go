package k8s

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

func newDisableCmd() *cobra.Command {
	disableCmd := &cobra.Command{
		Use:   "disable <component>",
		Short: "Disable a specific component in the cluster",
		Long:  fmt.Sprintf("Disable one of the specific components: %s.", strings.Join(componentList, ",")),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("Too many arguments. Please, only provide the name of the component that should be disabled.")
			}
			if len(args) < 1 {
				return fmt.Errorf("Not enough arguments. Please, provide the name of the component that should be disabled.")
			}
			if !slices.Contains(componentList, args[0]) {
				return fmt.Errorf("Unknown component %q. Needs to be one of: %s", args[0], strings.Join(componentList, ", "))
			}
			return nil
		},
	}

	disableCmd.AddCommand(newDisableDNSCmd())
	disableCmd.AddCommand(newDisableNetworkCmd())
	disableCmd.AddCommand(newDisableStorageCmd())
	disableCmd.AddCommand(newDisableIngressCmd())
	disableCmd.AddCommand(newDisableGatewayCmd())
	disableCmd.AddCommand(newDisableLoadBalancerCmd())
	disableCmd.AddCommand(newDisableMetricsServerCmd())
	return disableCmd
}
