package newtypes

type Network struct {
	PodCIDR     *string `json:"pod-cidr,omitempty"`
	ServiceCIDR *string `json:"service-cidr,omitempty"`
}
