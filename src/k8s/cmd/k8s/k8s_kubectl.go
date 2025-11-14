package k8s

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newKubectlCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		Short:              "Integrated Kubernetes kubectl client",
		DisableFlagParsing: true,
		PreRun:             chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			binary, err := exec.LookPath("kubectl")
			if err != nil {
				cmd.PrintErrln("Error: kubectl binary not found")
				env.Exit(1)
				return
			}

			kubeconfigEnvKey := "KUBECONFIG"
			adminKubeconfigPath := filepath.Join(env.Snap.KubernetesConfigDir(), "admin.conf")

			if !cmdutil.ExistsInEnviron(env.Environ, kubeconfigEnvKey) {
				if _, err := os.Stat(adminKubeconfigPath); err != nil {
					if os.IsNotExist(err) {
						cmd.PrintErrf("Error: %s file does not exist. Either set KUBECONFIG environment variable or ensure this node is bootstrapped as a control-plane node.\n", adminKubeconfigPath)
					} else {
						cmd.PrintErrf("Error: unable to access %s: %v\n", adminKubeconfigPath, err)
					}
					env.Exit(1)
					return
				}
			}

			command := append([]string{"kubectl"}, args...)
			environ := cmdutil.EnvironWithDefaults(
				env.Environ,
				kubeconfigEnvKey, adminKubeconfigPath,
				"EDITOR", "nano",
			)
			if err := syscall.Exec(binary, command, environ); err != nil {
				cmd.PrintErrf("Failed to run %s.\n\nThe error was: %v\n", command, err)
				env.Exit(1)
				return
			}
		},
	}
}
