package k8s

import (
	"syscall"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/spf13/cobra"
)

func newInspectCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	// We're copying the help string from the "inspect.sh" script with
	// only minor adjustments. At the same time, we'll avoid parsing the
	// same arguments twice.
	return &cobra.Command{
		Use:   "inspect <output-file>",
		Short: "Generate inspection report",
		Long: `
This command collects diagnostics and other relevant information from a Kubernetes
node (either control-plane or worker node) and compiles them into a tarball report.
The collected data includes service arguments, Kubernetes cluster info, SBOM, system
diagnostics, network diagnostics, and more. The command needs to be run with
elevated permissions (sudo).

Arguments:
  output-file             (Optional) The full path and filename for the generated tarball.
                          If not provided, a default filename based on the current date
                          and time will be used.
  --all-namespaces        (Optional) Acquire detailed debugging information, including logs
                          from all Kubernetes namespaces.
  --num-snap-log-entries  (Optional) The maximum number of log entries to collect
                          from snap services. Default: 100000.
  --timeout               (Optional) The maximum time in seconds to wait for a command.
                          Default: 180s.
`,
		DisableFlagParsing: true,
		PreRun:             chainPreRunHooks(hookRequireRoot(env)),
		Run: func(cmd *cobra.Command, args []string) {
			inspectScriptPath := env.Snap.K8sInspectScriptPath()

			command := append([]string{inspectScriptPath}, args...)
			environ := cmdutil.EnvironWithDefaults(
				env.Environ,
			)
			if err := syscall.Exec(inspectScriptPath, command, environ); err != nil {
				cmd.PrintErrf("Failed to run %s.\n\nError: %v\n", command, err)
				env.Exit(1)
				return
			}
		},
	}
}
