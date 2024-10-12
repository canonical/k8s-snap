package types

type Certificates struct {
	CACert                     *string `json:"ca-crt,omitempty"`
	CAKey                      *string `json:"ca-key,omitempty"`
	ClientCACert               *string `json:"client-ca-crt,omitempty"`
	ClientCAKey                *string `json:"client-ca-key,omitempty"`
	FrontProxyCACert           *string `json:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey            *string `json:"front-proxy-ca-key,omitempty"`
	ServiceAccountKey          *string `json:"service-account-key,omitempty"`
	APIServerKubeletClientCert *string `json:"apiserver-to-kubelet-client-crt,omitempty"`
	APIServerKubeletClientKey  *string `json:"apiserver-to-kubelet-client-key,omitempty"`
	AdminClientCert            *string `json:"admin-client-crt,omitempty"`
	AdminClientKey             *string `json:"admin-client-key,omitempty"`
	K8sdPublicKey              *string `json:"k8sd-public-key,omitempty"`
	K8sdPrivateKey             *string `json:"k8sd-private-key,omitempty"`
}

func (c Certificates) GetCACert() string { return getField(c.CACert) }
func (c Certificates) GetCAKey() string  { return getField(c.CAKey) }
func (c Certificates) GetClientCACert() string {
	// versions before 1.30.2 were using the same CA for server and client certificates
	if v := getField(c.ClientCACert); v != "" {
		return v
	}
	return c.GetCACert()
}
func (c Certificates) GetClientCAKey() string {
	// versions before 1.30.2 were using the same CA for server and client certificates
	if v := getField(c.ClientCAKey); v != "" {
		return v
	}
	return c.GetCAKey()
}
func (c Certificates) GetFrontProxyCACert() string  { return getField(c.FrontProxyCACert) }
func (c Certificates) GetFrontProxyCAKey() string   { return getField(c.FrontProxyCAKey) }
func (c Certificates) GetServiceAccountKey() string { return getField(c.ServiceAccountKey) }
func (c Certificates) GetAPIServerKubeletClientCert() string {
	return getField(c.APIServerKubeletClientCert)
}
func (c Certificates) GetAPIServerKubeletClientKey() string {
	return getField(c.APIServerKubeletClientKey)
}
func (c Certificates) GetAdminClientCert() string { return getField(c.AdminClientCert) }
func (c Certificates) GetAdminClientKey() string  { return getField(c.AdminClientKey) }
func (c Certificates) GetK8sdPublicKey() string   { return getField(c.K8sdPublicKey) }
func (c Certificates) GetK8sdPrivateKey() string  { return getField(c.K8sdPrivateKey) }

// Empty returns true if all Certificates fields are unset.
func (c Certificates) Empty() bool { return c == Certificates{} }
