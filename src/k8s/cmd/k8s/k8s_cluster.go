package k8s

import (
	"os"

	"github.com/canonical/k8s/pkg/snap"
)

var (
	clusterCmdOpts struct {
		stateDir string
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.stateDir, "state-dir", snap.NewSnap(os.Getenv("SNAP"), os.Getenv("SNAP_COMMON")).K8sdStateDir(), "Directory with the dqlite datastore")

	// By default, the state dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	rootCmd.Flags().MarkHidden("state-dir")
}
