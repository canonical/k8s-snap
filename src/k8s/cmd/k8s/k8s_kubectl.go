package k8s

import (
	"os/exec"
	"path"
	"syscall"

	cmdutil "github.com/canonical/k8s/cmd/util"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/spf13/cobra"
)

func newKubectlCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	return &cobra.Command{
		Use:                "kubectl",
		GroupID:            "general",
		Short:              "Integrated Kubernetes kubectl client",
		DisableFlagParsing: true,
		PreRun:             chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			isWorker, err := snaputil.IsWorker(env.Snap)
			if err != nil {
				cmd.PrintErrf("Error: Failed to check if this is worker-only node.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if isWorker {
				cmd.PrintErrln("Error: k8s kubectl commands are not allowed on worker nodes")
				env.Exit(1)
				return
			}

			client, err := env.Client(cmd.Context())
			if err != nil {
				cmd.PrintErrf("Error: Failed to create a k8sd client. Make sure that the k8sd service is running.\n\nThe error was: %v\n", err)
				env.Exit(1)
				return
			}

			if !client.IsBootstrapped(cmd.Context()) {
				cmd.PrintErrln("Error: The node is not part of a Kubernetes cluster. You can bootstrap a new cluster with:\n\n  sudo k8s bootstrap")
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
				"KUBECONFIG", path.Join(env.Snap.KubernetesConfigDir(), "admin.conf"),
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
