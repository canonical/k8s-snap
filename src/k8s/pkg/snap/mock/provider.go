package mock

import (
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/v2/microcluster"
)

type Provider struct {
	MicroClusterFn                     func() *microcluster.MicroCluster
	SnapFn                             func() snap.Snap
	NotifyUpdateNodeConfigControllerFn func()
	NotifyFeatureControllerFn          func(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool)
}

func (p *Provider) MicroCluster() *microcluster.MicroCluster {
	if p.MicroClusterFn != nil {
		return p.MicroClusterFn()
	}
	return nil
}

func (p *Provider) Snap() snap.Snap {
	if p.SnapFn != nil {
		return p.SnapFn()
	}
	return nil
}

func (p *Provider) NotifyUpdateNodeConfigController() {
	if p.NotifyUpdateNodeConfigControllerFn != nil {
		p.NotifyUpdateNodeConfigControllerFn()
	}
}

func (p *Provider) NotifyFeatureController(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns bool) {
	if p.NotifyFeatureControllerFn != nil {
		p.NotifyFeatureControllerFn(network, gateway, ingress, loadBalancer, localStorage, metricsServer, dns)
	}
}
