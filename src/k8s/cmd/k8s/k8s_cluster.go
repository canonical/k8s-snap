package k8s

import (
	"os"
	"path"
)

var (
	clusterCmdOpts struct {
		stateDir string
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.stateDir, "state-dir", path.Join(os.Getenv("SNAP_COMMON"), "/var/lib/k8sd/state"), "Directory with the dqlite datastore")

	// By default, the storage dir is set to a fixed directory in the snap.
	// This shouldn't be overwritten by the user.
	rootCmd.Flags().MarkHidden("state-dir")
}
