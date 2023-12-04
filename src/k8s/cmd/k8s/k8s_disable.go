package k8s

import (
	"github.com/canonical/k8s/pkg/k8s/component"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {

	disableCmd := &cobra.Command{
		Use:       "disable [component]",
		Short:     "Disable a specific component in the cluster",
		Long:      "Disable one of the specific components: cni, dns, gateway or ingress.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"cni", "dns", "gateway", "ingress"},
		RunE:      runDisableCmd(),
	}

	rootCmd.AddCommand(disableCmd)
}

func runDisableCmd() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		if err := component.DisableComponent(args[0]); err != nil {
			return err
		}
		logrus.Infof("Component %s disabled", args[0])
		return nil
	}
}
