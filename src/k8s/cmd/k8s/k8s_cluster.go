package k8s

var (
	clusterCmdOpts struct {
		address    string
		port       string
		storageDir string
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.address, "address", "", "IP Address of the cluster - required if not running on k8s node")
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.port, "port", "6444", "Port on which the REST-API is exposed")
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.storageDir, "storage-dir", "", "Directory with the dqlite datastore")
}
