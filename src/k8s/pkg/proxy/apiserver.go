package proxy

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/canonical/k8s/pkg/log"
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
	ctx = log.NewContext(ctx, log.FromContext(ctx).WithName("apiserver-proxy"))

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
		log.FromContext(ctx).Error(err, "Failed to start")
	}
	cancel()
}

func (p *APIServerProxy) watchForNewEndpoints(ctx context.Context, cancel func(), endpoints []string) {
	log := log.FromContext(ctx).WithValues("controller", "watchendpoints")
	if p.RefreshCh == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.RefreshCh:
		}

		// TODO: use k8s.GetKubernetesEndpoints instead
		newEndpoints, err := getKubernetesEndpoints(ctx, p.KubeconfigFile)
		switch {
		case err != nil:
			log.Error(err, "Failed to retrieve Kubernetes endpoints")
			continue
		case len(newEndpoints) == 0:
			log.Info("Warning: empty list of endpoints, skipping update")
			continue
		case len(newEndpoints) == len(endpoints) && reflect.DeepEqual(newEndpoints, endpoints):
			continue
		}
		log = log.WithValues("endpoints", endpoints)
		log.Info("Updating endpoints")

		if err := WriteEndpointsConfig(newEndpoints, p.EndpointsConfigFile); err != nil {
			log.Error(err, "Failed to update configuration file with new endpoints")
			continue
		}

		// cancel context in order to restart the proxy
		cancel()
		return
	}
}
