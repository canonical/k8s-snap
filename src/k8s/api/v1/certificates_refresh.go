package apiv1

type RefreshCertificatesPlanResponse struct {
	Seed                        int      `json:"seed"`
	CertificatesSigningRequests []string `json:"certificates_signing_requests"`
}

type RefreshCertificatesRunRequest struct {
	Seed              int `json:"seed"`
	ExpirationSeconds int `json:"expiration_seconds"`
}

type RefreshCertificatesRunResponse struct {
	ExpirationSeconds int `json:"expiration_seconds"`
}
