package k8s

var (
	clusterCmdOpts struct {
		address  string
		stateDir string
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.address, "address", "", "IP Address of the cluster - required if not running on k8s node")
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.stateDir, "state-dir", "", "Path to state store of local k8sd instance.")
}
