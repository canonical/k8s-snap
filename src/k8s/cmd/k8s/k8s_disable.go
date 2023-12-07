package k8s

import (
	"github.com/canonical/k8s/pkg/component"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {

	disableCmd := &cobra.Command{
		Use:       "disable <component>",
		Short:     "Disable a specific component in the cluster",
		Long:      "Disable one of the specific components: cni, dns, gateway, ingress, rbac or storage.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: componentList,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			var client component.ComponentManager
			client, err := component.NewManager()
			if err != nil {
				return err
			}

			if err := client.Disable(name); err != nil {
				return err
			}

			logrus.WithField("component", name).Info("Component disabled")
			return nil
		},
	}

	rootCmd.AddCommand(disableCmd)
}
