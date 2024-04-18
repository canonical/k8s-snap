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
	Config    UserFacingClusterConfig   `json:"config,omitempty" yaml:"config,omitempty"`
	Datastore UserFacingDatastoreConfig `json:"datastore,omitempty" yaml:"datastore,omitempty"`
}

type UpdateClusterConfigResponse struct {
}

type UserFacingClusterConfig struct {
	Network       NetworkConfig       `json:"network,omitempty" yaml:"network,omitempty"`
	DNS           DNSConfig           `json:"dns,omitempty" yaml:"dns,omitempty"`
	Ingress       IngressConfig       `json:"ingress,omitempty" yaml:"ingress,omitempty"`
	LoadBalancer  LoadBalancerConfig  `json:"load-balancer,omitempty" yaml:"load-balancer,omitempty"`
	LocalStorage  LocalStorageConfig  `json:"local-storage,omitempty" yaml:"local-storage,omitempty"`
	Gateway       GatewayConfig       `json:"gateway,omitempty" yaml:"gateway,omitempty"`
	MetricsServer MetricsServerConfig `json:"metrics-server,omitempty" yaml:"metrics-server,omitempty"`
	CloudProvider *string             `json:"cloud-provider,omitempty" yaml:"cloud-provider,omitempty"`
}

type DNSConfig struct {
	Enabled             *bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	ClusterDomain       *string   `json:"cluster-domain,omitempty" yaml:"cluster-domain,omitempty"`
	ServiceIP           *string   `json:"service-ip,omitempty" yaml:"service-ip,omitempty"`
	UpstreamNameservers *[]string `json:"upstream-nameservers,omitempty" yaml:"upstream-nameservers,omitempty"`
}

func (c DNSConfig) GetEnabled() bool                 { return getField(c.Enabled) }
func (c DNSConfig) GetClusterDomain() string         { return getField(c.ClusterDomain) }
func (c DNSConfig) GetServiceIP() string             { return getField(c.ServiceIP) }
func (c DNSConfig) GetUpstreamNameservers() []string { return getField(c.UpstreamNameservers) }

type IngressConfig struct {
	Enabled             *bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	DefaultTLSSecret    *string `json:"default-tls-secret,omitempty" yaml:"default-tls-secret,omitempty"`
	EnableProxyProtocol *bool   `json:"enable-proxy-protocol,omitempty" yaml:"enable-proxy-protocol,omitempty"`
}

func (c IngressConfig) GetEnabled() bool             { return getField(c.Enabled) }
func (c IngressConfig) GetDefaultTLSSecret() string  { return getField(c.DefaultTLSSecret) }
func (c IngressConfig) GetEnableProxyProtocol() bool { return getField(c.EnableProxyProtocol) }

type LoadBalancerConfig struct {
	Enabled        *bool     `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	CIDRs          *[]string `json:"cidrs,omitempty" yaml:"cidrs,omitempty"`
	L2Mode         *bool     `json:"l2-mode,omitempty" yaml:"l2-mode,omitempty"`
	L2Interfaces   *[]string `json:"l2-interfaces,omitempty" yaml:"l2-interfaces,omitempty"`
	BGPMode        *bool     `json:"bgp-mode,omitempty" yaml:"bgp-mode,omitempty"`
	BGPLocalASN    *int      `json:"bgp-local-asn,omitempty" yaml:"bgp-local-asn,omitempty"`
	BGPPeerAddress *string   `json:"bgp-peer-address,omitempty" yaml:"bgp-peer-address,omitempty"`
	BGPPeerASN     *int      `json:"bgp-peer-asn,omitempty" yaml:"bgp-peer-asn,omitempty"`
	BGPPeerPort    *int      `json:"bgp-peer-port,omitempty" yaml:"bgp-peer-port,omitempty"`
}

func (c LoadBalancerConfig) GetEnabled() bool          { return getField(c.Enabled) }
func (c LoadBalancerConfig) GetCIDRs() []string        { return getField(c.CIDRs) }
func (c LoadBalancerConfig) GetL2Mode() bool           { return getField(c.L2Mode) }
func (c LoadBalancerConfig) GetL2Interfaces() []string { return getField(c.L2Interfaces) }
func (c LoadBalancerConfig) GetBGPMode() bool          { return getField(c.BGPMode) }
func (c LoadBalancerConfig) GetBGPLocalASN() int       { return getField(c.BGPLocalASN) }
func (c LoadBalancerConfig) GetBGPPeerAddress() string { return getField(c.BGPPeerAddress) }
func (c LoadBalancerConfig) GetBGPPeerASN() int        { return getField(c.BGPPeerASN) }
func (c LoadBalancerConfig) GetBGPPeerPort() int       { return getField(c.BGPPeerPort) }

type LocalStorageConfig struct {
	Enabled       *bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	LocalPath     *string `json:"local-path,omitempty" yaml:"local-path,omitempty"`
	ReclaimPolicy *string `json:"reclaim-policy,omitempty" yaml:"reclaim-policy,omitempty"`
	Default       *bool   `json:"default,omitempty" yaml:"default,omitempty"`
}

func (c LocalStorageConfig) GetEnabled() bool         { return getField(c.Enabled) }
func (c LocalStorageConfig) GetLocalPath() string     { return getField(c.LocalPath) }
func (c LocalStorageConfig) GetReclaimPolicy() string { return getField(c.ReclaimPolicy) }
func (c LocalStorageConfig) GetDefault() bool         { return getField(c.Default) }

type NetworkConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

func (c NetworkConfig) GetEnabled() bool { return getField(c.Enabled) }

type GatewayConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

func (c GatewayConfig) GetEnabled() bool { return getField(c.Enabled) }

type MetricsServerConfig struct {
	Enabled *bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

func (c MetricsServerConfig) GetEnabled() bool { return getField(c.Enabled) }

type UserFacingDatastoreConfig struct {
	// Type of the datastore. Needs to be "external".
	Type       *string   `json:"type,omitempty" yaml:"type,omitempty"`
	Servers    *[]string `json:"servers,omitempty" yaml:"servers,omitempty"`
	CACert     *string   `json:"ca-crt,omitempty" yaml:"ca-crt,omitempty"`
	ClientCert *string   `json:"client-crt,omitempty" yaml:"client-crt,omitempty"`
	ClientKey  *string   `json:"client-key,omitempty" yaml:"client-key,omitempty"`
}

func (c UserFacingDatastoreConfig) GetType() string       { return getField(c.Type) }
func (c UserFacingDatastoreConfig) GetServers() []string  { return getField(c.Servers) }
func (c UserFacingDatastoreConfig) GetCACert() string     { return getField(c.CACert) }
func (c UserFacingDatastoreConfig) GetClientCert() string { return getField(c.ClientCert) }
func (c UserFacingDatastoreConfig) GetClientKey() string  { return getField(c.ClientKey) }

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
