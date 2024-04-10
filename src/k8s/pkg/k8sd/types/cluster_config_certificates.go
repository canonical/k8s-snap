package types

type Certificates struct {
	CACert                     *string `json:"ca-crt,omitempty"`
	CAKey                      *string `json:"ca-key,omitempty"`
	FrontProxyCACert           *string `json:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey            *string `json:"front-proxy-ca-key,omitempty"`
	ServiceAccountKey          *string `json:"service-account-key,omitempty"`
	APIServerKubeletClientCert *string `json:"apiserver-to-kubelet-client-crt,omitempty"`
	APIServerKubeletClientKey  *string `json:"apiserver-to-kubelet-client-key,omitempty"`
}

func (c Certificates) GetCACert() string            { return getField(c.CACert) }
func (c Certificates) GetCAKey() string             { return getField(c.CAKey) }
func (c Certificates) GetFrontProxyCACert() string  { return getField(c.FrontProxyCACert) }
func (c Certificates) GetFrontProxyCAKey() string   { return getField(c.FrontProxyCAKey) }
func (c Certificates) GetServiceAccountKey() string { return getField(c.ServiceAccountKey) }
func (c Certificates) GetAPIServerKubeletClientCert() string {
	return getField(c.APIServerKubeletClientCert)
}
func (c Certificates) GetAPIServerKubeletClientKey() string {
	return getField(c.APIServerKubeletClientKey)
}

// Empty returns true if all Certificates fields are unset
func (c Certificates) Empty() bool {
	return c.CACert == nil && c.CAKey == nil && c.FrontProxyCACert == nil && c.FrontProxyCAKey == nil && c.ServiceAccountKey == nil && c.APIServerKubeletClientCert == nil && c.APIServerKubeletClientKey == nil
}
