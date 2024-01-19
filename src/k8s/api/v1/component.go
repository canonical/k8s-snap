package v1

// GetComponentsRequest is used to list components info.
type GetComponentsRequest struct{}

// GetComponentResponse is the response for "GET 1.0/k8sd/components".
type GetComponentsResponse struct {
	Components []Component `json:"components"`
}

// UpdateComponentRequest is used to update component state.
type UpdateComponentRequest struct {
	Status ComponentStatus `json:"status"`
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

// UpdateComponentResponse is the response for "PUT 1.0/k8sd/components/dns".
type UpdateDNSComponentResponse struct{}

// UpdateComponentResponse is the response for "PUT 1.0/k8sd/components".
type UpdateComponentResponse struct{}

// Component holds information about a k8s component.
type Component struct {
	Name   string          `json:"name"`
	Status ComponentStatus `json:"status"`
}

type ComponentStatus string

const (
	Unknown          ComponentStatus = "unknown"
	ComponentEnable  ComponentStatus = "enable"
	ComponentDisable ComponentStatus = "disable"
)
