package k8s_apiserver_proxy

import (
	"time"

	cmdutil "github.com/canonical/k8s/cmd/util"
	"github.com/canonical/k8s/pkg/proxy"
	"github.com/spf13/cobra"
)

func NewRootCmd(env cmdutil.ExecutionEnvironment) *cobra.Command {
	var opts struct {
		listenAddress              string
		endpointsConfigFile        string
		refreshEndpointsInterval   time.Duration
		refreshEndpointsKubeconfig string
	}

	cmd := &cobra.Command{
		Use:   "k8s-apiserver-proxy",
		Short: "Local API server proxy used in worker nodes. Forwards requests to active kube-apiserver instances",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set input/output streams
			cmd.SetIn(env.Stdin)
			cmd.SetOut(env.Stdout)
			cmd.SetErr(env.Stderr)

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			var refreshCh <-chan time.Time
			if opts.refreshEndpointsInterval == 0 {
				cmd.Println("Will not auto-refresh list of control plane endpoints")
			} else {
				if opts.refreshEndpointsInterval < 15*time.Second {
					cmd.PrintErrf("Refresh interval %v is less than minimum of 15s. Using the minimum 15s instead.\n", opts.refreshEndpointsInterval)
					opts.refreshEndpointsInterval = 15 * time.Second
				}
				refreshCh = time.NewTicker(opts.refreshEndpointsInterval).C
			}

			p := &proxy.APIServerProxy{
				ListenAddress:       opts.listenAddress,
				EndpointsConfigFile: opts.endpointsConfigFile,
				KubeconfigFile:      opts.refreshEndpointsKubeconfig,
				RefreshCh:           refreshCh,
			}

			if err := p.Run(cmd.Context()); err != nil {
				cmd.PrintErrf("Proxy failed with error: %v", err)
				env.Exit(1)
				return
			}
		},
	}

	cmd.Flags().StringVar(&opts.listenAddress, "listen", ":6443", "listen address")
	cmd.Flags().StringVar(&opts.endpointsConfigFile, "endpoints", "/etc/kubernetes/k8s-apiserver-proxy.json", "configuration file with known kube-apiserver endpoints")
	cmd.Flags().StringVar(&opts.refreshEndpointsKubeconfig, "kubeconfig", "/etc/kubernetes/kubelet.conf", "kubeconfig file to use for updating list of known kube-apiserver endpoints")
	cmd.Flags().DurationVar(&opts.refreshEndpointsInterval, "refresh-interval", 30*time.Second, "interval between checking for new kube-apiserver endpoints. set to 0 to disable")

	return cmd
}
