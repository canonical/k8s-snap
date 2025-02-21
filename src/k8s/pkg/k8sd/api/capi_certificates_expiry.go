package api

import (
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
)

func (e *Endpoints) postCertificatesExpiry(s state.State, r *http.Request) response.Response {
	config, err := databaseutil.GetClusterConfig(r.Context(), s)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get cluster config: %w", err))
	}

	certificates := []string{
		config.Certificates.GetCACert(),
		config.Certificates.GetClientCACert(),
		config.Certificates.GetAdminClientCert(),
		config.Certificates.GetAPIServerKubeletClientCert(),
		config.Certificates.GetFrontProxyCACert(),
	}

	var earliestExpiry time.Time
	// Find the earliest expiry certificate
	// They should all be about the same but better double-check this.
	for _, cert := range certificates {
		if cert == "" {
			continue
		}

		cert, _, err := pkiutil.LoadCertificate(cert, "")
		if err != nil {
			return response.InternalError(fmt.Errorf("failed to load certificate: %w", err))
		}

		if earliestExpiry.IsZero() || cert.NotAfter.Before(earliestExpiry) {
			earliestExpiry = cert.NotAfter
		}
	}

	return response.SyncResponse(true, &apiv1.CertificatesExpiryResponse{
		ExpiryDate: earliestExpiry.Format(time.RFC3339),
	})
}
