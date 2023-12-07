package k8s

import (
	"github.com/canonical/k8s/pkg/component"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var componentList = []string{"cni", "dns", "gateway", "ingress", "rbac", "storage"}

func init() {

	enableCmd := &cobra.Command{
		Use:       "enable <component>",
		Short:     "Enable a specific component in the cluster",
		Long:      "Enable one of the specific components: cni, dns, gateway, ingress, rbac or storage.",
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: componentList,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			var client component.ComponentManager
			client, err := component.NewManager()
			if err != nil {
				return err
			}

			err = client.Enable(name)
			if err != nil {
				return err
			}

			logrus.WithField("component", name).Info("Component enabled")
			return nil

		},
	}

	rootCmd.AddCommand(enableCmd)
}
