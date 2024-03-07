package k8s

import (
	"github.com/canonical/k8s/pkg/utils/shims"
	"github.com/spf13/cobra"
)

var xPrintShimPidsCmd = &cobra.Command{
	Use:    "x-print-shim-pids",
	Short:  "Print list of PIDs started by the containerd shim and pause processes",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		pids, err := shims.RunningContainerdShimPIDs(cmd.Context())
		if err != nil {
			panic(err)
		}
		for _, pid := range pids {
			cmd.Println(pid)
		}
	},
}
