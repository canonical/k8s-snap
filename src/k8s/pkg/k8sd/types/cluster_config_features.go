package types

type DNS struct {
	Enabled             *bool     `json:"enabled,omitempty"`
	UpstreamNameservers *[]string `json:"upstream-nameservers,omitempty"`
}

type Ingress struct {
	Enabled             *bool   `json:"enabled,omitempty"`
	DefaultTLSSecret    *string `json:"default-tls-secret,omitempty"`
	EnableProxyProtocol *bool   `json:"enable-proxy-protocol,omitempty"`
}

type LoadBalancer struct {
	Enabled        *bool                   `json:"enabled,omitempty"`
	CIDRs          *[]string               `json:"cidrs,omitempty"`
	IPRanges       *[]LoadBalancer_IPRange `json:"ranges,omitempty"`
	L2Mode         *bool                   `json:"l2-mode,omitempty"`
	L2Interfaces   *[]string               `json:"l2-interfaces,omitempty"`
	BGPMode        *bool                   `json:"bgp-mode,omitempty"`
	BGPLocalASN    *int                    `json:"bgp-local-asn,omitempty"`
	BGPPeerAddress *string                 `json:"bgp-peer-address,omitempty"`
	BGPPeerASN     *int                    `json:"bgp-peer-asn,omitempty"`
	BGPPeerPort    *int                    `json:"bgp-peer-port,omitempty"`
}

type LoadBalancer_IPRange struct {
	Start string `json:"start"`
	Stop  string `json:"stop"`
}

type Gateway struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type MetricsServer struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type LocalStorage struct {
	Enabled       *bool   `json:"enabled,omitempty"`
	LocalPath     *string `json:"local-path,omitempty"`
	ReclaimPolicy *string `json:"reclaim-policy,omitempty"`
	Default       *bool   `json:"default,omitempty"`
}

func (c DNS) GetEnabled() bool                 { return getField(c.Enabled) }
func (c DNS) GetUpstreamNameservers() []string { return getField(c.UpstreamNameservers) }
func (c DNS) Empty() bool                      { return c.Enabled == nil && c.UpstreamNameservers == nil }

func (c Ingress) GetEnabled() bool             { return getField(c.Enabled) }
func (c Ingress) GetDefaultTLSSecret() string  { return getField(c.DefaultTLSSecret) }
func (c Ingress) GetEnableProxyProtocol() bool { return getField(c.EnableProxyProtocol) }
func (c Ingress) Empty() bool {
	return c.Enabled == nil && c.DefaultTLSSecret == nil && c.EnableProxyProtocol == nil
}

func (c Gateway) GetEnabled() bool { return getField(c.Enabled) }
func (c Gateway) Empty() bool      { return c.Enabled == nil }

func (c LoadBalancer) GetEnabled() bool                    { return getField(c.Enabled) }
func (c LoadBalancer) GetCIDRs() []string                  { return getField(c.CIDRs) }
func (c LoadBalancer) GetIPRanges() []LoadBalancer_IPRange { return getField(c.IPRanges) }
func (c LoadBalancer) GetL2Mode() bool                     { return getField(c.L2Mode) }
func (c LoadBalancer) GetL2Interfaces() []string           { return getField(c.L2Interfaces) }
func (c LoadBalancer) GetBGPMode() bool                    { return getField(c.BGPMode) }
func (c LoadBalancer) GetBGPLocalASN() int                 { return getField(c.BGPLocalASN) }
func (c LoadBalancer) GetBGPPeerAddress() string           { return getField(c.BGPPeerAddress) }
func (c LoadBalancer) GetBGPPeerASN() int                  { return getField(c.BGPPeerASN) }
func (c LoadBalancer) GetBGPPeerPort() int                 { return getField(c.BGPPeerPort) }
func (c LoadBalancer) Empty() bool {
	return c.Enabled == nil && c.CIDRs == nil && c.L2Mode == nil && c.L2Interfaces == nil && c.BGPMode == nil && c.BGPLocalASN == nil && c.BGPPeerAddress == nil && c.BGPPeerASN == nil && c.BGPPeerPort == nil
}

func (c LocalStorage) GetEnabled() bool         { return getField(c.Enabled) }
func (c LocalStorage) GetLocalPath() string     { return getField(c.LocalPath) }
func (c LocalStorage) GetReclaimPolicy() string { return getField(c.ReclaimPolicy) }
func (c LocalStorage) GetDefault() bool         { return getField(c.Default) }
func (c LocalStorage) Empty() bool {
	return c.Enabled == nil && c.LocalPath == nil && c.ReclaimPolicy == nil && c.Default == nil
}

func (c MetricsServer) GetEnabled() bool { return getField(c.Enabled) }
func (c MetricsServer) Empty() bool      { return c.Enabled == nil }
