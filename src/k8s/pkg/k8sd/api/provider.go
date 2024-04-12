package api

import (
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/microcluster/microcluster"
)

// Provider is an interface for state that the API endpoints need access to.
type Provider interface {
	MicroCluster() *microcluster.MicroCluster
	Snap() snap.Snap
	UpdateNodeConfigurationControllerCh() chan<- struct{}
}
