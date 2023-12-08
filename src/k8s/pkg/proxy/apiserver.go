package proxy

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"
)

// APIServerProxy is a TCP proxy that forwards requests to the API Servers of the cluster.
type APIServerProxy struct {
	// ListenAddress is the address where the proxy will accept connections.
	ListenAddress string

	// EndpointsConfigFile is the config file with the initial kube-apiserver endpoints.
	EndpointsConfigFile string

	// RefreshCh signals the proxy to update the list of known kube-apiserver endpoints. If the list
	// of kube-apiserver endpoints have changed, the endpoints config file is updated, and the proxy
	// will restart automatically.
	RefreshCh <-chan time.Time

	// Kubeconfig is the kubeconfig file to use to refresh the kube-apiserver endpoints.
	KubeconfigFile string
}

// Run starts the proxy.
func (p *APIServerProxy) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		cfg, err := loadEndpointsConfig(p.EndpointsConfigFile)
		if err != nil {
			return fmt.Errorf("failed to load endpoints configuration: %w", err)
		}

		proxyCtx, cancel := context.WithCancel(ctx)
		go p.startProxy(proxyCtx, cancel, cfg.Endpoints)
		go p.watchForNewEndpoints(proxyCtx, cancel, cfg.Endpoints)
		<-proxyCtx.Done()
	}
}

func (p *APIServerProxy) startProxy(ctx context.Context, cancel func(), endpoints []string) {
	if err := startProxy(ctx, p.ListenAddress, endpoints); err != nil {
		log.Println(fmt.Errorf("apiserver proxy failed: %w", err))
	}
	cancel()
}

func (p *APIServerProxy) watchForNewEndpoints(ctx context.Context, cancel func(), endpoints []string) {
	if p.RefreshCh == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.RefreshCh:
		}

		newEndpoints, err := getKubernetesEndpoints(ctx, p.KubeconfigFile)
		switch {
		case err != nil:
			log.Println(fmt.Errorf("failed to retrieve kubernetes endpoints: %w", err))
			continue
		case len(newEndpoints) == 0:
			log.Println("warning: empty list of endpoints, skipping update")
			continue
		case len(newEndpoints) == len(endpoints) && reflect.DeepEqual(newEndpoints, endpoints):
			continue
		}
		log.Println("updating endpoints")

		if err := writeEndpointsConfig(newEndpoints, p.EndpointsConfigFile); err != nil {
			log.Printf("failed to update configuration file with new endpoints: %s", err)
			continue
		}

		// cancel context in order to restart the proxy
		cancel()
		return
	}
}
