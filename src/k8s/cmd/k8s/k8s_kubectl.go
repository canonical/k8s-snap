package k8s

import (
	"os/exec"
	"path/filepath"
	"syscall"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
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
			client, err := env.Snap.K8sdClient("")
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if response, initialized, err := client.NodeStatus(cmd.Context()); err != nil {
				cmd.PrintErrf("Error: Failed to retrieve the node status.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			} else if !initialized {
				cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
				env.Exit(1)
				return
			} else if response.NodeStatus.ClusterRole == apiv1.ClusterRoleWorker {
				cmd.PrintErrln("Error: k8s kubectl commands are not allowed on worker nodes.")
				env.Exit(1)
				return
			}

			binary, err := exec.LookPath("kubectl")
			if err != nil {
				cmd.PrintErrln("Error: kubectl binary not found")
				env.Exit(1)
				return
			}

			command := append([]string{"kubectl"}, args...)
			environ := cmdutil.EnvironWithDefaults(
				env.Environ,
				"KUBECONFIG", filepath.Join(env.Snap.KubernetesConfigDir(), "admin.conf"),
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
