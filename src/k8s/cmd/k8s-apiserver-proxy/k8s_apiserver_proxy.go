package k8s_apiserver_proxy

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/canonical/k8s/pkg/proxy"
	"github.com/spf13/cobra"
)

var (
	rootCmdOpts struct {
		listenAddress              string
		endpointsConfigFile        string
		refreshEndpointsInterval   time.Duration
		refreshEndpointsKubeconfig string
	}

	rootCmd = &cobra.Command{
		Use:   "k8s-apiserver-proxy",
		Short: "Local API server proxy used in worker nodes. Forwards requests to active kube-apiserver instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			var refreshCh <-chan time.Time
			if rootCmdOpts.refreshEndpointsInterval == 0 {
				log.Println("Will not auto-refresh list of control plane endpoints")
			} else {
				if rootCmdOpts.refreshEndpointsInterval < 15*time.Second {
					log.Printf("Refresh interval %v is less than minimum of 15s. Using the minimum 15s instead.\n", rootCmdOpts.refreshEndpointsInterval)
					rootCmdOpts.refreshEndpointsInterval = 15 * time.Second
				}
				refreshCh = time.NewTicker(rootCmdOpts.refreshEndpointsInterval).C
			}

			p := &proxy.APIServerProxy{
				ListenAddress:       rootCmdOpts.listenAddress,
				EndpointsConfigFile: rootCmdOpts.endpointsConfigFile,
				KubeconfigFile:      rootCmdOpts.refreshEndpointsKubeconfig,
				RefreshCh:           refreshCh,
			}

			ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			if err := p.Run(ctx); err != nil {
				return fmt.Errorf("proxy failed: %w", err)
			}
			return nil
		},
	}
)

func init() {
	rootCmd.Flags().StringVar(&rootCmdOpts.listenAddress, "listen", ":6443", "listen address")
	rootCmd.Flags().StringVar(&rootCmdOpts.endpointsConfigFile, "endpoints", "/etc/kubernetes/k8s-apiserver-proxy.json", "configuration file with known kube-apiserver endpoints")
	rootCmd.Flags().StringVar(&rootCmdOpts.refreshEndpointsKubeconfig, "kubeconfig", "/etc/kubernetes/kubelet.conf", "kubeconfig file to use for updating list of known kube-apiserver endpoints")
	rootCmd.Flags().DurationVar(&rootCmdOpts.refreshEndpointsInterval, "refresh-interval", 30*time.Second, "interval between checking for new kube-apiserver endpoints. set to 0 to disable")
}
