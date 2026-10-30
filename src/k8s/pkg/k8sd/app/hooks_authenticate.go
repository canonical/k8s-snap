package app

import (
	"context"
	"crypto/subtle"
	"crypto/x509"
	"errors"
	"net/http"
	"time"

	"github.com/canonical/k8sd/pkg/log"
	mctypes "github.com/canonical/microcluster/v3/microcluster/types"
)

// onAuthenticate verifies that the request's client certificate is valid and present in the truststore.
func (a *App) onAuthenticate(ctx context.Context, s mctypes.State, r *http.Request, trustedCerts map[string]x509.Certificate) (bool, error) {
	if r.TLS == nil {
		return false, errors.New("no TLS connection state")
	}

	if len(r.TLS.PeerCertificates) == 0 {
		return false, errors.New("no certificate present in request")
	}

	peerCertificate := r.TLS.PeerCertificates[0]

	now := time.Now()
	if now.Before(peerCertificate.NotBefore) {
		return false, errors.New("certificate is not yet valid")
	}

	if now.After(peerCertificate.NotAfter) {
		return false, errors.New("certificate has expired")
	}

	log := log.FromContext(ctx).WithValues("hook", "authenticate")

	for fingerprint, v := range trustedCerts {
		if subtle.ConstantTimeCompare(peerCertificate.Raw, v.Raw) == 1 {
			log.V(1).Info("Matched trusted cert", "fingerprint", fingerprint, "subject", v.Subject.String())
			return true, nil
		}
	}

	return false, nil
}
