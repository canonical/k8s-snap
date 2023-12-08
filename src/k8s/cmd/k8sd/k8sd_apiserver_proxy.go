package k8sd

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
	apiServerProxyCmdOpts struct {
		listenAddress              string
		endpointsConfigFile        string
		refreshEndpointsInterval   time.Duration
		refreshEndpointsKubeconfig string
	}

	apiServerProxyCmd = &cobra.Command{
		Use:    "apiserver-proxy",
		Short:  "Local API server proxy used in worker nodes. Forwards requests to the active kube-apiserver instances",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var refreshCh <-chan time.Time
			if apiServerProxyCmdOpts.refreshEndpointsInterval == 0 {
				log.Println("Will not auto-refresh list of control plane endpoints")
			} else {
				if apiServerProxyCmdOpts.refreshEndpointsInterval < 15*time.Second {
					log.Printf("Refresh interval %v is less than minimum of 15s. Using the minimum 15s instead.\n", apiServerProxyCmdOpts.refreshEndpointsInterval)
					apiServerProxyCmdOpts.refreshEndpointsInterval = 15 * time.Second
				}
				refreshCh = time.NewTicker(apiServerProxyCmdOpts.refreshEndpointsInterval).C
			}

			p := &proxy.APIServerProxy{
				ListenAddress:       apiServerProxyCmdOpts.listenAddress,
				EndpointsConfigFile: apiServerProxyCmdOpts.endpointsConfigFile,
				KubeconfigFile:      apiServerProxyCmdOpts.refreshEndpointsKubeconfig,
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
	apiServerProxyCmd.Flags().StringVar(&apiServerProxyCmdOpts.listenAddress, "listen", ":6443", "listen address")
	apiServerProxyCmd.Flags().StringVar(&apiServerProxyCmdOpts.endpointsConfigFile, "endpoints", "/etc/kubernetes/apiserver-proxy-config.json", "configuration file with known kube-apiserver endpoints")
	apiServerProxyCmd.Flags().StringVar(&apiServerProxyCmdOpts.refreshEndpointsKubeconfig, "kubeconfig", "/etc/kubernetes/kubelet.config", "kubeconfig file to use for updating list of known kube-apiserver endpoints")
	apiServerProxyCmd.Flags().DurationVar(&apiServerProxyCmdOpts.refreshEndpointsInterval, "refresh-interval", 30*time.Second, "interval between checking for new kube-apiserver endpoints. set to 0 to disable")

	rootCmd.AddCommand(apiServerProxyCmd)
}
