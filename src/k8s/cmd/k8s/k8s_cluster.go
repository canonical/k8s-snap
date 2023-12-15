package k8s

import (
	"strconv"

	"github.com/canonical/k8s/pkg/config"
	"github.com/canonical/k8s/pkg/snap"
)

var (
	clusterCmdOpts struct {
		remoteAddress string
		port          string
		storageDir    string
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.remoteAddress, "remote-address", "", "IP Address of another cluster member")
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.port, "port", strconv.Itoa(config.DefaultPort), "Port on which the REST-API is exposed")
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.storageDir, "storage-dir", snap.CommonPath("/var/lib/k8sd"), "Directory with the dqlite datastore")
}
