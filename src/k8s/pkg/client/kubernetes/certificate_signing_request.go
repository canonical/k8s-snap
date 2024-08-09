package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/log"
	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WatchCertificateSigningRequest watches a CertificateSigningRequest with the
// given name and calls a verify function on each event.
// WatchCertificateSigningRequest will continue watching the CSR until the
// verify function returns true or an non-retriable error occurs.
//
// The verify function should return true if the CSR is valid and processing
// should stop.
// The verify function should return false if the CSR is not yet valid and
// processing should continue.
// The verify function should return an error if the CSR is in an invalid state
// (e.g., failed or denied) or the issued certificate is invalid.
func (c *Client) WatchCertificateSigningRequest(ctx context.Context, name string, verify func(csr *certificatesv1.CertificateSigningRequest) (bool, error)) error {
	log := log.FromContext(ctx)
	for {
		w, err := c.CertificatesV1().CertificateSigningRequests().Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
		if err != nil {
			log.V(1).Info("Failed to watch CSR", "error", err)
			continue
		}
		watchClosed := false
		for !watchClosed {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case evt, ok := <-w.ResultChan():
				if !ok {
					log.V(1).Info("Watch closed")
					watchClosed = true
					continue
				}

				csr, ok := evt.Object.(*certificatesv1.CertificateSigningRequest)
				if !ok {
					log.V(1).Info("Expected a CertificateSigningRequest but received something else", "object", evt.Object)
					watchClosed = true
					continue
				}

				if valid, err := verify(csr); err != nil {
					return fmt.Errorf("failed to verify CSR %s: %w", name, err)
				} else if valid {
					return nil
				}

			}
		}

		w.Stop()
		log.V(1).Info("Retrying to watch CSR")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}
