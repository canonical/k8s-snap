package kubernetes

import (
	"context"
	"fmt"

	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WatchCertificateSigningRequest watches a CertificateSigningRequest with the
// given name and calls a verify function on each event.
// WatchCertificateSigningRequest will continue watching the CSR until the
// verify function returns true or an non-retriable error occurs.
// WatchCertificateSigningRequest will return true and a wrapped error if the
// error is retriable.
// WatchCertificateSigningRequest will return false and a wrapped error if the
// error is not retriable.
//
// The verify function should return true if the CSR is valid and processing
// should stop.
// The verify function should return false if the CSR is not yet valid and
// processing should continue.
// The verify function should return an error if the CSR is in an invalid state
// (e.g., failed or denied) or the issued certificate is invalid.
func (c *Client) WatchCertificateSigningRequest(ctx context.Context, name string, verify func(csr *certificatesv1.CertificateSigningRequest) (bool, error)) (bool, error) {
	w, err := c.CertificatesV1().CertificateSigningRequests().Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return true, fmt.Errorf("failed to watch CSR %s: %w", name, err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case evt, ok := <-w.ResultChan():
			if !ok {
				return true, fmt.Errorf("watch closed")
			}

			csr, ok := evt.Object.(*certificatesv1.CertificateSigningRequest)
			if !ok {
				return true, fmt.Errorf("expected a CertificateSigningRequest but received %#v", evt.Object)
			}

			if valid, err := verify(csr); err != nil {
				return false, fmt.Errorf("failed to verify CSR %s: %w", name, err)
			} else if valid {
				return false, nil
			}

		}
	}
}
