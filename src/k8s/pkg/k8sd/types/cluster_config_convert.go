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
			config.Features.Network.Enabled = vals.Pointer(true)
		case "dns":
			config.Features.DNS.Enabled = vals.Pointer(true)
		case "local-storage":
			config.Features.LocalStorage.Enabled = vals.Pointer(true)
		case "ingress":
			config.Features.Ingress.Enabled = vals.Pointer(true)
		case "gateway":
			config.Features.Gateway.Enabled = vals.Pointer(true)
		case "metrics-server":
			config.Features.MetricsServer.Enabled = vals.Pointer(true)
		case "load-balancer":
			config.Features.LoadBalancer.Enabled = vals.Pointer(true)
		}
	}

	return config
}

// ClusterConfigFromUserFacing converts UserFacingClusterConfig from public API into a ClusterConfig.
func ClusterConfigFromUserFacing(u *apiv1.UserFacingClusterConfig) ClusterConfig {
	return ClusterConfig{
		Kubelet: Kubelet{
			ClusterDNS:    u.DNS.ServiceIP,
			ClusterDomain: u.DNS.ClusterDomain,
		},
		Features: Features{
			Network: NetworkFeature{
				Enabled: u.Network.Enabled,
			},
			DNS: DNSFeature{
				Enabled:             u.DNS.Enabled,
				UpstreamNameservers: u.DNS.UpstreamNameservers,
			},
			Ingress: IngressFeature{
				Enabled:             u.Ingress.Enabled,
				DefaultTLSSecret:    u.Ingress.DefaultTLSSecret,
				EnableProxyProtocol: u.Ingress.EnableProxyProtocol,
			},
			LoadBalancer: LoadBalancerFeature{
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
			LocalStorage: LocalStorageFeature{
				Enabled:       u.LocalStorage.Enabled,
				LocalPath:     u.LocalStorage.LocalPath,
				ReclaimPolicy: u.LocalStorage.ReclaimPolicy,
				SetDefault:    u.LocalStorage.SetDefault,
			},
			MetricsServer: MetricsServerFeature{
				Enabled: u.MetricsServer.Enabled,
			},
			Gateway: GatewayFeature{
				Enabled: u.Gateway.Enabled,
			},
		},
	}
}

// ClusterConfigToUserFacing converts a ClusterConfig to a UserFacingClusterConfig from the public API.
func ClusterConfigToUserFacing(c ClusterConfig) apiv1.UserFacingClusterConfig {
	return apiv1.UserFacingClusterConfig{
		Network: apiv1.NetworkConfig{
			Enabled: c.Features.Network.Enabled,
		},
		DNS: apiv1.DNSConfig{
			Enabled:             c.Features.DNS.Enabled,
			ClusterDomain:       c.Kubelet.ClusterDomain,
			ServiceIP:           c.Kubelet.ClusterDNS,
			UpstreamNameservers: c.Features.DNS.UpstreamNameservers,
		},
		Ingress: apiv1.IngressConfig{
			Enabled:             c.Features.Ingress.Enabled,
			DefaultTLSSecret:    c.Features.Ingress.DefaultTLSSecret,
			EnableProxyProtocol: c.Features.Ingress.EnableProxyProtocol,
		},
		LoadBalancer: apiv1.LoadBalancerConfig{
			Enabled:        c.Features.LoadBalancer.Enabled,
			CIDRs:          c.Features.LoadBalancer.CIDRs,
			L2Mode:         c.Features.LoadBalancer.L2Mode,
			L2Interfaces:   c.Features.LoadBalancer.L2Interfaces,
			BGPMode:        c.Features.LoadBalancer.BGPMode,
			BGPLocalASN:    c.Features.LoadBalancer.BGPLocalASN,
			BGPPeerAddress: c.Features.LoadBalancer.BGPPeerAddress,
			BGPPeerASN:     c.Features.LoadBalancer.BGPPeerASN,
			BGPPeerPort:    c.Features.LoadBalancer.BGPPeerPort,
		},
		LocalStorage: apiv1.LocalStorageConfig{
			Enabled:       c.Features.LocalStorage.Enabled,
			LocalPath:     c.Features.LocalStorage.LocalPath,
			ReclaimPolicy: c.Features.LocalStorage.ReclaimPolicy,
			SetDefault:    c.Features.LocalStorage.SetDefault,
		},
		MetricsServer: apiv1.MetricsServerConfig{
			Enabled: c.Features.MetricsServer.Enabled,
		},
		Gateway: apiv1.GatewayConfig{
			Enabled: c.Features.Gateway.Enabled,
		},
	}
}
