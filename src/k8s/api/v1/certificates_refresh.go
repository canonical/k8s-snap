package apiv1

type RefreshCertificatesPlanResponse struct {
	Seed                        int      `json:"seed"`
	CertificatesSigningRequests []string `json:"certificates-signing-requests"`
}

type RefreshCertificatesRunRequest struct {
	Seed              int `json:"seed"`
	ExpirationSeconds int `json:"expiration-seconds"`
}

type RefreshCertificatesRunResponse struct {
	ExpirationSeconds int `json:"expiration-seconds"`
}
