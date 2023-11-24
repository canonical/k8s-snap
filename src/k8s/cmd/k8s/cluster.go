package k8s

var (
	clusterCmdOpts struct {
		flagAddress  string
		flagStateDir string
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.flagAddress, "address", "", "IP Address of the cluster - required if not running on k8s node"+"``")
	rootCmd.PersistentFlags().StringVar(&clusterCmdOpts.flagStateDir, "state-dir", "", "Path to state store of local k8sd instance.")
}
