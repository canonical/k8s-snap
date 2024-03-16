package newtypes

type ContainerdRegistry struct {
	Host         string   `json:"host,omitempty"`
	URLs         []string `json:"urls,omitempty"`
	Username     string   `json:"username,omitempty"`
	Password     string   `json:"password,omitempty"`
	Token        string   `json:"token,omitempty"`
	OverridePath bool     `json:"override-path,omitempty"`
	SkipVerify   bool     `json:"skip-verify,omitempty"`
	// TODO(neoaggelos): add option to configure certificates for containerd registries
	// CACert       string
	// ClientCert   string
	// ClientKey    string
}

type Containerd struct {
	Registries *[]ContainerdRegistry `json:"registries,omitempty"`
}
