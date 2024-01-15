package k8s

import (
	"fmt"
	"os"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	kubectlCmd = &cobra.Command{
		Use:   "kubectl",
		Short: "Integrated Kubernetes CLI",
		// All commands should be passed to kubectl
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if rootCmdOpts.logDebug {
				logrus.SetLevel(logrus.TraceLevel)
			}

			// Allow users to provide their own kubeconfig but
			// fallback to the admin config if nothing is provided.
			if os.Getenv("KUBECONFIG") == "" {
				os.Setenv("KUBECONFIG", utils.SnapCommonPath("/etc/kubernetes/admin.conf"))
			}
			command := append(
				[]string{utils.SnapPath("bin/kubectl"),
					"--kubeconfig",
					os.Getenv("KUBECONFIG"),
				},
				args...,
			)

			err := utils.RunCommand(cmd.Context(), command...)
			if err != nil {
				fmt.Println(err)
				return fmt.Errorf("failed to execute kubectl command: %w", err)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(kubectlCmd)
}
