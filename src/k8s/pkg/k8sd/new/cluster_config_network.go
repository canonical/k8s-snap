package newtypes

type Network struct {
	PodCIDR     *string `json:"pod-cidr,omitempty"`
	ServiceCIDR *string `json:"service-cidr,omitempty"`
}

func (c Network) GetPodCIDR() string     { return getField(c.PodCIDR) }
func (c Network) GetServiceCIDR() string { return getField(c.ServiceCIDR) }
