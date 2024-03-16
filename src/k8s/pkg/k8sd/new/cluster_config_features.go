package newtypes

type NetworkFeature struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type DNSFeature struct {
	Enabled             *bool     `json:"enabled,omitempty"`
	UpstreamNameservers *[]string `json:"upstream-nameservers,omitempty"`
}

type IngressFeature struct {
	Enabled             *bool   `json:"enabled,omitempty"`
	DefaultTLSSecret    *string `json:"default-tls-secret,omitempty"`
	EnableProxyProtocol *bool   `json:"enable-proxy-protocol,omitempty"`
}

type LoadBalancerFeature struct {
	Enabled        *bool     `json:"enabled,omitempty"`
	CIDRs          *[]string `json:"cidrs,omitempty"`
	L2Mode         *bool     `json:"l2-mode,omitempty"`
	L2Interfaces   *[]string `json:"l2-interfaces,omitempty"`
	BGPMode        *bool     `json:"bgp-mode,omitempty"`
	BGPLocalASN    *int      `json:"bgp-local-asn,omitempty"`
	BGPPeerAddress *string   `json:"bgp-peer-address,omitempty"`
	BGPPeerASN     *int      `json:"bgp-peer-asn,omitempty"`
	BGPPeerPort    *int      `json:"bgp-peer-port,omitempty"`
}

type GatewayFeature struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type MetricsServerFeature struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type LocalStorageFeature struct {
	Enabled       *bool   `json:"enabled,omitempty"`
	LocalPath     *string `json:"local-path,omitempty"`
	ReclaimPolicy *string `json:"reclaim-policy,omitempty"`
	SetDefault    *bool   `json:"set-default,omitempty"`
}

type Features struct {
	Network       NetworkFeature       `json:"network,omitempty"`
	DNS           DNSFeature           `json:"dns,omitempty"`
	Ingress       IngressFeature       `json:"ingress,omitempty"`
	LoadBalancer  LoadBalancerFeature  `json:"load-balancer,omitempty"`
	Gateway       GatewayFeature       `json:"gateway,omitempty"`
	LocalStorage  LocalStorageFeature  `json:"local-storage,omitempty"`
	MetricsServer MetricsServerFeature `json:"metrics-server,omitempty"`
}

func (c NetworkFeature) GetEnabled() bool { return getField(c.Enabled) }

func (c DNSFeature) GetEnabled() bool                 { return getField(c.Enabled) }
func (c DNSFeature) GetUpstreamNameservers() []string { return getField(c.UpstreamNameservers) }

func (c IngressFeature) GetEnabled() bool             { return getField(c.Enabled) }
func (c IngressFeature) GetDefaultTLSSecret() string  { return getField(c.DefaultTLSSecret) }
func (c IngressFeature) GetEnableProxyProtocol() bool { return getField(c.EnableProxyProtocol) }

func (c GatewayFeature) GetEnabled() bool { return getField(c.Enabled) }

func (c LoadBalancerFeature) GetEnabled() bool          { return getField(c.Enabled) }
func (c LoadBalancerFeature) GetCIDRs() []string        { return getField(c.CIDRs) }
func (c LoadBalancerFeature) GetL2Mode() bool           { return getField(c.L2Mode) }
func (c LoadBalancerFeature) GetL2Interfaces() []string { return getField(c.L2Interfaces) }
func (c LoadBalancerFeature) GetBGPMode() bool          { return getField(c.BGPMode) }
func (c LoadBalancerFeature) GetBGPLocalASN() int       { return getField(c.BGPLocalASN) }
func (c LoadBalancerFeature) GetBGPPeerAddress() string { return getField(c.BGPPeerAddress) }
func (c LoadBalancerFeature) GetBGPPeerASN() int        { return getField(c.BGPPeerASN) }
func (c LoadBalancerFeature) GetBGPPeerPort() int       { return getField(c.BGPPeerPort) }

func (c LocalStorageFeature) GetEnabled() bool         { return getField(c.Enabled) }
func (c LocalStorageFeature) GetLocalPath() string     { return getField(c.LocalPath) }
func (c LocalStorageFeature) GetReclaimPolicy() string { return getField(c.ReclaimPolicy) }
func (c LocalStorageFeature) GetSetDefault() bool      { return getField(c.Enabled) }

func (c MetricsServerFeature) GetEnabled() bool { return getField(c.Enabled) }
