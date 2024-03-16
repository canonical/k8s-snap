package newtypes

import (
	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
)

// ClusterConfigFromBootstrapConfig extracts the cluster config parts from the BootstrapConfig
// and maps them to a ClusterConfig.
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
