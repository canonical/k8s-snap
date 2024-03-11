package types

import (
	"fmt"
	"net"
	"strings"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils/vals"
)

// ClusterConfig is the control plane configuration format of the k8s cluster.
// ClusterConfig should attempt to use structured fields wherever possible.
type ClusterConfig struct {
	Network       Network       `yaml:"network"`
	Certificates  Certificates  `yaml:"certificates"`
	Kubelet       Kubelet       `yaml:"kubelet"`
	K8sDqlite     K8sDqlite     `yaml:"k8s-dqlite"`
	APIServer     APIServer     `yaml:"apiserver"`
	DNS           DNS           `yaml:"dns"`
	Ingress       Ingress       `yaml:"ingress"`
	LoadBalancer  LoadBalancer  `yaml:"load-balancer"`
	LocalStorage  LocalStorage  `yaml:"local-storage"`
	Gateway       Gateway       `yaml:"gateway"`
	MetricsServer MetricsServer `yaml:"metrics-server"`
	Containerd    Containerd    `yaml:"containerd"`
}

type Network struct {
	Enabled     *bool  `yaml:"enabled,omitempty"`
	PodCIDR     string `yaml:"pod-cidr,omitempty"`
	ServiceCIDR string `yaml:"svc-cidr,omitempty"`
}

type Certificates struct {
	CACert                     string `yaml:"ca-crt,omitempty"`
	CAKey                      string `yaml:"ca-key,omitempty"`
	APIServerKubeletClientCert string `yaml:"apiserver-kubelet-client-crt,omitempty"`
	APIServerKubeletClientKey  string `yaml:"apiserver-kubelet-client-key,omitempty"`
	K8sDqliteCert              string `yaml:"k8s-dqlite-crt,omitempty"`
	K8sDqliteKey               string `yaml:"k8s-dqlite-key,omitempty"`
	FrontProxyCACert           string `yaml:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey            string `yaml:"front-proxy-ca-key,omitempty"`
}

type Kubelet struct {
	CloudProvider string `yaml:"cloud-provider,omitempty"`
	ClusterDNS    string `yaml:"cluster-dns,omitempty"`
	ClusterDomain string `yaml:"cluster-domain,omitempty"`
}

type APIServer struct {
	SecurePort          int    `yaml:"secure-port,omitempty"`
	AuthorizationMode   string `yaml:"authorization-mode,omitempty"`
	ServiceAccountKey   string `yaml:"service-account-key,omitempty"`
	Datastore           string `yaml:"datastore,omitempty"`
	DatastoreURL        string `yaml:"datastore-url,omitempty"`
	DatastoreCA         string `yaml:"datastore-ca-crt,omitempty"`
	DatastoreClientCert string `yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey  string `yaml:"datastore-client-key,omitempty"`
}

type K8sDqlite struct {
	Port int `yaml:"port,omitempty"`
}

type DNS struct {
	Enabled             *bool    `yaml:"enabled,omitempty"`
	UpstreamNameservers []string `yaml:"upstream-nameservers,omitempty"`
}

type Ingress struct {
	Enabled             *bool  `yaml:"enabled,omitempty"`
	DefaultTLSSecret    string `yaml:"default-tls-secret,omitempty"`
	EnableProxyProtocol *bool  `yaml:"enable-proxy-protocol,omitempty"`
}

type LoadBalancer struct {
	Enabled        *bool    `yaml:"enabled,omitempty"`
	CIDRs          []string `yaml:"cidrs,omitempty"`
	L2Enabled      *bool    `yaml:"l2-enabled,omitempty"`
	L2Interfaces   []string `yaml:"l2-interfaces,omitempty"`
	BGPEnabled     *bool    `yaml:"bgp-enabled,omitempty"`
	BGPLocalASN    int      `yaml:"bgp-local-asn,omitempty"`
	BGPPeerAddress string   `yaml:"bgp-peer-address,omitempty"`
	BGPPeerASN     int      `yaml:"bgp-peer-asn,omitempty"`
	BGPPeerPort    int      `yaml:"bgp-peer-port,omitempty"`
}

type LocalStorage struct {
	Enabled       *bool  `yaml:"enabled,omitempty"`
	LocalPath     string `yaml:"local-path,omitempty"`
	ReclaimPolicy string `yaml:"reclaim-policy,omitempty"`
	SetDefault    *bool  `yaml:"set-default,omitempty"`
}

type Gateway struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

type MetricsServer struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

type Containerd struct {
	Registries []ContainerdRegistry `yaml:"registries,omitempty"`
}

type ContainerdRegistry struct {
	Host         string   `yaml:"host"`
	URLs         []string `yaml:"urls"`
	Username     string   `yaml:"username,omitempty"`
	Password     string   `yaml:"password,omitempty"`
	Token        string   `yaml:"token,omitempty"`
	OverridePath bool     `yaml:"overridePath,omitempty"`
	SkipVerify   bool     `yaml:"skipVerify,omitempty"`
	// TODO(neoaggelos): add option to configure certificates for containerd registries
	// CA           string   `yaml:"ca,omitempty"`
	// Cert         string   `yaml:"cert,omitempty"`
	// Key          string   `yaml:"key,omitempty"`
}

func (c *ClusterConfig) Validate() error {
	clusterCIDRs := strings.Split(c.Network.PodCIDR, ",")
	if len(clusterCIDRs) != 1 && len(clusterCIDRs) != 2 {
		return fmt.Errorf("invalid number of cluster CIDRs: %d", len(clusterCIDRs))
	}
	serviceCIDRs := strings.Split(c.Network.ServiceCIDR, ",")
	if len(serviceCIDRs) != 1 && len(serviceCIDRs) != 2 {
		return fmt.Errorf("invalid number of service CIDRs: %d", len(serviceCIDRs))
	}

	for _, cidr := range append(clusterCIDRs, serviceCIDRs...) {
		_, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return fmt.Errorf("invalid CIDR: %w", err)
		}
	}

	return nil
}

func (c *ClusterConfig) SetDefaults() {
	if c.Network.PodCIDR == "" {
		c.Network.PodCIDR = "10.1.0.0/16"
	}
	if c.Network.ServiceCIDR == "" {
		c.Network.ServiceCIDR = "10.152.183.0/24"
	}
	if c.APIServer.Datastore == "" {
		c.APIServer.Datastore = "k8s-dqlite"
	}
	if c.APIServer.SecurePort == 0 {
		c.APIServer.SecurePort = 6443
	}
	if c.APIServer.AuthorizationMode == "" {
		c.APIServer.AuthorizationMode = "Node,RBAC"
	}
	if c.K8sDqlite.Port == 0 {
		c.K8sDqlite.Port = 9000
	}
	if c.DNS.UpstreamNameservers == nil {
		c.DNS.UpstreamNameservers = []string{"/etc/resolv.conf"}
	}
	if c.Kubelet.ClusterDomain == "" {
		c.Kubelet.ClusterDomain = "cluster.local"
	}
	if c.LocalStorage.LocalPath == "" {
		c.LocalStorage.LocalPath = "/var/snap/k8s/common/rawfile-storage"
	}
	if c.LocalStorage.ReclaimPolicy == "" {
		c.LocalStorage.ReclaimPolicy = "Delete"
	}
	if c.LocalStorage.SetDefault == nil {
		c.LocalStorage.SetDefault = vals.Pointer(true)
	}
	if c.LoadBalancer.L2Enabled == nil {
		c.LoadBalancer.L2Enabled = vals.Pointer(true)
	}
}

// ClusterConfigFromBootstrapConfig extracts the cluster config parts from the BootstrapConfig
// and maps them to a ClusterConfig.
func ClusterConfigFromBootstrapConfig(b *apiv1.BootstrapConfig) ClusterConfig {
	authzMode := "Node,RBAC"
	// Only disable rbac if explicitly set to false during bootstrap
	if v := b.EnableRBAC; v != nil && !*v {
		authzMode = "AlwaysAllow"
	}

	config := ClusterConfig{
		APIServer: APIServer{
			AuthorizationMode: authzMode,
		},
		Network: Network{
			PodCIDR:     b.ClusterCIDR,
			ServiceCIDR: b.ServiceCIDR,
		},
		K8sDqlite: K8sDqlite{
			Port: b.K8sDqlitePort,
		},
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

func ClusterConfigFromUserFacing(ufConfig *apiv1.UserFacingClusterConfig) ClusterConfig {
	config := ClusterConfig{}

	if ufConfig.DNS != nil {
		config.Kubelet = Kubelet{
			ClusterDNS:    ufConfig.DNS.ServiceIP,
			ClusterDomain: ufConfig.DNS.ClusterDomain,
		}

		config.DNS = DNS{
			Enabled:             ufConfig.DNS.Enabled,
			UpstreamNameservers: ufConfig.DNS.UpstreamNameservers,
		}
	}

	if ufConfig.Network != nil {
		config.Network = Network{
			Enabled: ufConfig.Network.Enabled,
		}
	}

	if ufConfig.Ingress != nil {
		config.Ingress = Ingress{
			Enabled:             ufConfig.Ingress.Enabled,
			DefaultTLSSecret:    ufConfig.Ingress.DefaultTLSSecret,
			EnableProxyProtocol: ufConfig.Ingress.EnableProxyProtocol,
		}
	}

	if ufConfig.LoadBalancer != nil {
		// TODO(berkayoz): make sure everything about bgp to be set if bgp enabled
		config.LoadBalancer = LoadBalancer{
			Enabled:        ufConfig.LoadBalancer.Enabled,
			CIDRs:          ufConfig.LoadBalancer.CIDRs,
			L2Enabled:      ufConfig.LoadBalancer.L2Enabled,
			L2Interfaces:   ufConfig.LoadBalancer.L2Interfaces,
			BGPEnabled:     ufConfig.LoadBalancer.BGPEnabled,
			BGPLocalASN:    ufConfig.LoadBalancer.BGPLocalASN,
			BGPPeerAddress: ufConfig.LoadBalancer.BGPPeerAddress,
			BGPPeerASN:     ufConfig.LoadBalancer.BGPPeerASN,
			BGPPeerPort:    ufConfig.LoadBalancer.BGPPeerPort,
		}
	}

	if ufConfig.LocalStorage != nil {
		config.LocalStorage = LocalStorage{
			Enabled:       ufConfig.LocalStorage.Enabled,
			LocalPath:     ufConfig.LocalStorage.LocalPath,
			ReclaimPolicy: ufConfig.LocalStorage.ReclaimPolicy,
			SetDefault:    ufConfig.LocalStorage.SetDefault,
		}
	}

	if ufConfig.Gateway != nil {
		config.Gateway = Gateway{
			Enabled: ufConfig.Gateway.Enabled,
		}
	}

	if ufConfig.MetricsServer != nil {
		config.MetricsServer = MetricsServer{
			Enabled: ufConfig.MetricsServer.Enabled,
		}
	}

	return config
}
