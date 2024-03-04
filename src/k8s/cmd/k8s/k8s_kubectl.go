package k8s

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/cmd/k8s/errors"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/spf13/cobra"
)

func newKubectlCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kubectl",
		Short: "Integrated Kubernetes kubectl client",
		// All commands should be passed to kubectl
		DisableFlagParsing: true,
		PreRunE:            chainPreRunHooks(hookSetupClient),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			defer errors.Transform(&err, nil)

			snap := snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_COMMON"))

			isWorker, err := snaputil.IsWorker(snap)
			if err != nil {
				return fmt.Errorf("failed to check if node is a worker: %w", err)
			}

			if isWorker {
				// TODO: convert to sentinel errors to work with human readability machinery #130
				return fmt.Errorf("this action is restricted on workers")
			}

			if !k8sdClient.IsBootstrapped(cmd.Context()) {
				return apiv1.ErrNotBootstrapped
			}

			// Allow users to provide their own kubeconfig but
			// fallback to the admin config if nothing is provided.
			if os.Getenv("KUBECONFIG") == "" {
				os.Setenv("KUBECONFIG", "/etc/kubernetes/admin.conf")
			}
			// Set a default editor that comes with the snap so that 'kubectl edit' works
			if os.Getenv("EDITOR") == "" {
				os.Setenv("EDITOR", "nano")
			}
			path, err := exec.LookPath("kubectl")
			if err != nil {
				return fmt.Errorf("kubectl not found")
			}

			command := append(
				[]string{"kubectl"},
				args...,
			)
			// completly replace the executable with kubectl
			// as we want to be as close as possible to a "real"
			// kubectl invocation.
			return syscall.Exec(path, command, os.Environ())
		},
	}
}
