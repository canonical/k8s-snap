package types

import (
	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
)

// ClusterConfigFromBootstrapConfig converts BootstrapConfig from public API into a ClusterConfig.
func ClusterConfigFromBootstrapConfig(b *apiv1.BootstrapConfig) ClusterConfig {
	var config ClusterConfig

	authorizationMode := "Node,RBAC"
	if !vals.OptionalBool(b.EnableRBAC, true) {
		authorizationMode = "AlwaysAllow"
	}
	config.APIServer.AuthorizationMode = vals.Pointer(authorizationMode)

	switch b.Datastore {
	case "", "k8s-dqlite":
		config.Datastore = Datastore{
			Type:          vals.Pointer("k8s-dqlite"),
			K8sDqlitePort: vals.Pointer(b.K8sDqlitePort),
		}
	case "external":
		config.Datastore = Datastore{
			Type:               vals.Pointer("external"),
			ExternalURL:        vals.Pointer(b.DatastoreURL),
			ExternalCACert:     vals.Pointer(b.DatastoreCACert),
			ExternalClientCert: vals.Pointer(b.DatastoreClientCert),
			ExternalClientKey:  vals.Pointer(b.DatastoreClientKey),
		}
	}

	if b.ClusterCIDR != "" {
		config.Network.PodCIDR = vals.Pointer(b.ClusterCIDR)
	}
	if b.ServiceCIDR != "" {
		config.Network.ServiceCIDR = vals.Pointer(b.ServiceCIDR)
	}

	for _, component := range b.Components {
		switch component {
		case "network":
			config.Network.Enabled = vals.Pointer(true)
		case "dns":
			config.DNS.Enabled = vals.Pointer(true)
		case "local-storage":
			config.LocalStorage.Enabled = vals.Pointer(true)
		case "ingress":
			config.Ingress.Enabled = vals.Pointer(true)
		case "gateway":
			config.Gateway.Enabled = vals.Pointer(true)
		case "metrics-server":
			config.MetricsServer.Enabled = vals.Pointer(true)
		case "load-balancer":
			config.LoadBalancer.Enabled = vals.Pointer(true)
		}
	}

	return config
}

// ClusterConfigFromUserFacing converts UserFacingClusterConfig from public API into a ClusterConfig.
func ClusterConfigFromUserFacing(u apiv1.UserFacingClusterConfig) ClusterConfig {
	return ClusterConfig{
		Kubelet: Kubelet{
			ClusterDNS:    u.DNS.ServiceIP,
			ClusterDomain: u.DNS.ClusterDomain,
		},
		Network: Network{
			Enabled: u.Network.Enabled,
		},
		DNS: DNS{
			Enabled:             u.DNS.Enabled,
			UpstreamNameservers: u.DNS.UpstreamNameservers,
		},
		Ingress: Ingress{
			Enabled:             u.Ingress.Enabled,
			DefaultTLSSecret:    u.Ingress.DefaultTLSSecret,
			EnableProxyProtocol: u.Ingress.EnableProxyProtocol,
		},
		LoadBalancer: LoadBalancer{
			Enabled:        u.LoadBalancer.Enabled,
			CIDRs:          u.LoadBalancer.CIDRs,
			L2Mode:         u.LoadBalancer.L2Mode,
			L2Interfaces:   u.LoadBalancer.L2Interfaces,
			BGPMode:        u.LoadBalancer.BGPMode,
			BGPLocalASN:    u.LoadBalancer.BGPLocalASN,
			BGPPeerAddress: u.LoadBalancer.BGPPeerAddress,
			BGPPeerASN:     u.LoadBalancer.BGPPeerASN,
			BGPPeerPort:    u.LoadBalancer.BGPPeerPort,
		},
		LocalStorage: LocalStorage{
			Enabled:       u.LocalStorage.Enabled,
			LocalPath:     u.LocalStorage.LocalPath,
			ReclaimPolicy: u.LocalStorage.ReclaimPolicy,
			SetDefault:    u.LocalStorage.SetDefault,
		},
		MetricsServer: MetricsServer{
			Enabled: u.MetricsServer.Enabled,
		},
		Gateway: Gateway{
			Enabled: u.Gateway.Enabled,
		},
	}
}

// ToUserFacing converts a ClusterConfig to a UserFacingClusterConfig from the public API.
func (c ClusterConfig) ToUserFacing() apiv1.UserFacingClusterConfig {
	return apiv1.UserFacingClusterConfig{
		Network: apiv1.NetworkConfig{
			Enabled: c.Network.Enabled,
		},
		DNS: apiv1.DNSConfig{
			Enabled:             c.DNS.Enabled,
			ClusterDomain:       c.Kubelet.ClusterDomain,
			ServiceIP:           c.Kubelet.ClusterDNS,
			UpstreamNameservers: c.DNS.UpstreamNameservers,
		},
		Ingress: apiv1.IngressConfig{
			Enabled:             c.Ingress.Enabled,
			DefaultTLSSecret:    c.Ingress.DefaultTLSSecret,
			EnableProxyProtocol: c.Ingress.EnableProxyProtocol,
		},
		LoadBalancer: apiv1.LoadBalancerConfig{
			Enabled:        c.LoadBalancer.Enabled,
			CIDRs:          c.LoadBalancer.CIDRs,
			L2Mode:         c.LoadBalancer.L2Mode,
			L2Interfaces:   c.LoadBalancer.L2Interfaces,
			BGPMode:        c.LoadBalancer.BGPMode,
			BGPLocalASN:    c.LoadBalancer.BGPLocalASN,
			BGPPeerAddress: c.LoadBalancer.BGPPeerAddress,
			BGPPeerASN:     c.LoadBalancer.BGPPeerASN,
			BGPPeerPort:    c.LoadBalancer.BGPPeerPort,
		},
		LocalStorage: apiv1.LocalStorageConfig{
			Enabled:       c.LocalStorage.Enabled,
			LocalPath:     c.LocalStorage.LocalPath,
			ReclaimPolicy: c.LocalStorage.ReclaimPolicy,
			SetDefault:    c.LocalStorage.SetDefault,
		},
		MetricsServer: apiv1.MetricsServerConfig{
			Enabled: c.MetricsServer.Enabled,
		},
		Gateway: apiv1.GatewayConfig{
			Enabled: c.Gateway.Enabled,
		},
	}
}
