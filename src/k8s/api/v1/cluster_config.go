package v1

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type GetClusterConfigRequest struct{}

type GetClusterConfigResponse struct {
	Config UserFacingClusterConfig
}

type UpdateClusterConfigRequest struct {
	Config UserFacingClusterConfig
}

type UpdateClusterConfigResponse struct {
}

type UserFacingClusterConfig struct {
	Network       *NetworkConfig       `json:"network,omitempty" yaml:"network,omitempty"`
	DNS           *DNSConfig           `json:"dns,omitempty" yaml:"dns,omitempty"`
	Ingress       *IngressConfig       `json:"ingress,omitempty" yaml:"ingress,omitempty"`
	LoadBalancer  *LoadBalancerConfig  `json:"load-balancer,omitempty" yaml:"load-balancer,omitempty"`
	LocalStorage  *LocalStorageConfig  `json:"local-storage,omitempty" yaml:"local-storage,omitempty"`
	Gateway       *GatewayConfig       `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	MetricsServer *MetricsServerConfig `json:"metrics-server,omitempty" yaml:"metrics-server,omitempty"`
	APIServer     *APIServerConfig     `json:"apiserver,omitempty" yaml:"apiserver,omitempty"`
}

type DNSConfig struct {
	Enabled             *bool    `json:"enabled,omitempty" yaml:"enabled"`
	ClusterDomain       string   `json:"cluster-domain,omitempty" yaml:"cluster-domain"`
	ServiceIP           string   `json:"service-ip,omitempty" yaml:"service-ip"`
	UpstreamNameservers []string `json:"upstream-nameservers,omitempty" yaml:"upstream-nameservers"`
}

type IngressConfig struct {
	Enabled             *bool  `json:"enabled,omitempty" yaml:"enabled"`
	DefaultTLSSecret    string `json:"default-tls-secret,omitempty" yaml:"default-tls-secret"`
	EnableProxyProtocol *bool  `json:"enable-proxy-protocol,omitempty" yaml:"enable-proxy-protocol"`
}

type LoadBalancerConfig struct {
	Enabled        *bool    `json:"enabled,omitempty" yaml:"enabled"`
	CIDRs          []string `json:"cidrs,omitempty" yaml:"cidrs"`
	L2Enabled      *bool    `json:"l2-mode,omitempty" yaml:"l2-mode"`
	L2Interfaces   []string `json:"l2-interfaces,omitempty" yaml:"l2-interfaces"`
	BGPEnabled     *bool    `json:"bgp-mode,omitempty" yaml:"bgp-mode"`
	BGPLocalASN    int      `json:"bgp-local-asn,omitempty" yaml:"bgp-local-asn"`
	BGPPeerAddress string   `json:"bgp-peer-address,omitempty" yaml:"bgp-peer-address"`
	BGPPeerASN     int      `json:"bgp-peer-asn,omitempty" yaml:"bgp-peer-asn"`
	BGPPeerPort    int      `json:"bgp-peer-port,omitempty" yaml:"bgp-peer-port"`
}

type LocalStorageConfig struct {
	Enabled       *bool  `json:"enabled,omitempty" yaml:"enabled"`
	LocalPath     string `json:"local-path,omitempty" yaml:"local-path"`
	ReclaimPolicy string `json:"reclaim-policy,omitempty" yaml:"reclaim-policy"`
	SetDefault    *bool  `json:"set-default,omitempty" yaml:"set-default"`
}

type NetworkConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled"`
}

type APIServerConfig struct {
	Datastore    string `json:"datastore,omitempty" yaml:"datastore"`
	DatastoreURL string `json:"datastore-url,omitempty" yaml:"datastore-url"`
}

type GatewayConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled"`
}

type MetricsServerConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled"`
}

func (c UserFacingClusterConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c NetworkConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c APIServerConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c DNSConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c IngressConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c LoadBalancerConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c LocalStorageConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c GatewayConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}

func (c MetricsServerConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("%#v\n", c)
	}
	return string(b)
}
