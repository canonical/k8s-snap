package types

import "net"

const (
	// ControlPlaneEndpointBackendExternal is the default backend. k8sd only publishes the
	// endpoint host in the kube-apiserver serving-certificate SANs; the operator owns the
	// load balancer that fronts the control-plane nodes.
	ControlPlaneEndpointBackendExternal = "external"
	// ControlPlaneEndpointBackendService realises the endpoint with an in-cluster
	// LoadBalancer Service (fronted by MetalLB) maintained by k8sd.
	ControlPlaneEndpointBackendService = "service"
)

// ControlPlaneEndpoint is the persisted form of the bootstrap control-plane endpoint.
type ControlPlaneEndpoint struct {
	Host    *string `json:"host,omitempty"`
	Port    *int    `json:"port,omitempty"`
	Backend *string `json:"backend,omitempty"`
}

func (c ControlPlaneEndpoint) GetHost() string    { return getField(c.Host) }
func (c ControlPlaneEndpoint) GetPort() int       { return getField(c.Port) }
func (c ControlPlaneEndpoint) GetBackend() string { return getField(c.Backend) }
func (c ControlPlaneEndpoint) Empty() bool        { return c == ControlPlaneEndpoint{} }

// SANs returns the endpoint host as either an IP SAN or a DNS SAN (exactly one of the two
// slices is non-empty), or zero values when no host is configured. It is used to inject the
// endpoint into the kube-apiserver serving-certificate at bootstrap, on join, and on every
// certificate rotation.
func (c ControlPlaneEndpoint) SANs() (ips []net.IP, dnsNames []string) {
	host := c.GetHost()
	if host == "" {
		return nil, nil
	}
	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}
	return nil, []string{host}
}
