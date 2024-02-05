package v1

// GetComponentsRequest is used to list components info.
type GetComponentsRequest struct{}

// GetComponentResponse is the response for "GET 1.0/k8sd/components".
type GetComponentsResponse struct {
	Components []Component `json:"components"`
}

// UpdateDNSComponentRequest is used to update the DNS component state.
type UpdateDNSComponentRequest struct {
	Status ComponentStatus    `json:"status"`
	Config DNSComponentConfig `json:"config,omitempty"`
}

// DNSComponentConfig holds the configuration values for the DNS component.
type DNSComponentConfig struct {
	ClusterDomain       string   `json:"clusterDomain,omitempty"`
	ServiceIP           string   `json:"serviceIP,omitempty"`
	UpstreamNameservers []string `json:"upstreamNameservers,omitempty"`
}

// UpdateNetworkComponentRequest is used to update the Network component state.
type UpdateNetworkComponentRequest struct {
	Status ComponentStatus `json:"status"`
}

// UpdateStorageComponentRequest is used to update the Storage component state.
type UpdateStorageComponentRequest struct {
	Status ComponentStatus `json:"status"`
}

// UpdateIngressComponentRequest is used to update the Ingress component state.
type UpdateIngressComponentRequest struct {
	Status ComponentStatus        `json:"status"`
	Config IngressComponentConfig `json:"config,omitempty"`
}

// IngressComponentConfig holds the configuration values for the Ingress component.
type IngressComponentConfig struct {
	DefaultTLSSecret    string `json:"defaultTLSSecret,omitempty"`
	EnableProxyProtocol bool   `json:"enableProxyProtocol,omitempty"`
}

// UpdateGatewayComponentRequest is used to update the Storage gateway state.
type UpdateGatewayComponentRequest struct {
	Status ComponentStatus `json:"status"`
}

// UpdateLoadBalancerComponentRequest is used to update the LoadBalancer component state.
type UpdateLoadBalancerComponentRequest struct {
	Status ComponentStatus             `json:"status"`
	Config LoadBalancerComponentConfig `json:"config,omitempty"`
}

// LoadBalancerComponentConfig holds the configuration values for the LoadBalancer component.
type LoadBalancerComponentConfig struct {
	CIDRs          []string `json:"cidrs,omitempty"`
	L2Enabled      bool     `json:"l2Enabled,omitempty"`
	L2Interfaces   []string `json:"l2Interfaces,omitempty"`
	BGPEnabled     bool     `json:"bgpEnabled,omitempty"`
	BGPLocalASN    int      `json:"bgpLocalAsn,omitempty"`
	BGPPeerAddress string   `json:"bgpPeerAddress,omitempty"`
	BGPPeerASN     int      `json:"bgpPeerAsn,omitempty"`
	BGPPeerPort    int      `json:"bgpPeerPort,omitempty"`
}

// UpdateDNSComponentResponse is the response for "PUT 1.0/k8sd/components/dns".
type UpdateDNSComponentResponse struct{}

// UpdateNetworkComponentResponse is the response for "PUT 1.0/k8sd/components/network".
type UpdateNetworkComponentResponse struct{}

// UpdateStorageComponentResponse is the response for "PUT 1.0/k8sd/components/storage".
type UpdateStorageComponentResponse struct{}

// UpdateIngressComponentResponse is the response for "PUT 1.0/k8sd/components/ingress".
type UpdateIngressComponentResponse struct{}

// UpdateGatewayComponentResponse is the response for "PUT 1.0/k8sd/components/gateway".
type UpdateGatewayComponentResponse struct{}

// UpdateLoadBalancerComponentResponse is the response for "PUT 1.0/k8sd/components/loadbalancer".
type UpdateLoadBalancerComponentResponse struct{}

// Component holds information about a k8s component.
type Component struct {
	Name   string          `json:"name"`
	Status ComponentStatus `json:"status"`
}

type ComponentStatus string

const (
	Unknown          ComponentStatus = "unknown"
	ComponentEnable  ComponentStatus = "enabled"
	ComponentDisable ComponentStatus = "disabled"
)
