package newtypes

type Kubelet struct {
	CloudProvider *string `json:"cloud-provider,omitempty"`
	ClusterDNS    *string `json:"cluster-dns,omitempty"`
	ClusterDomain *string `json:"cluster-domain,omitempty"`
}

func (c Kubelet) GetCloudProvider() string { return getField(c.CloudProvider) }
func (c Kubelet) GetClusterDNS() string    { return getField(c.ClusterDNS) }
func (c Kubelet) GetClusterDomain() string { return getField(c.ClusterDomain) }
func (c Kubelet) Empty() bool {
	return c.CloudProvider == nil && c.ClusterDNS == nil && c.ClusterDomain == nil
}
