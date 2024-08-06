package kubernetes

import (
	"context"
	"fmt"

	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WatchCertificateSigningRequest watches a CertificateSigningRequest with the given name and calls the verify function on each event.
func (c *Client) WatchCertificateSigningRequest(ctx context.Context, name string, verify func(csr *certificatesv1.CertificateSigningRequest) (bool, error)) error {

	w, err := c.CertificatesV1().CertificateSigningRequests().Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return fmt.Errorf("failed to watch CSR %s: %w", name, err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-w.ResultChan():
			if !ok {
				return fmt.Errorf("watch closed")
			}

			csr, ok := evt.Object.(*certificatesv1.CertificateSigningRequest)
			if !ok {
				return fmt.Errorf("expected a CertificateSigningRequest but received %#v", evt.Object)
			}

			valid, err := verify(csr)
			// If the verify function returns an error, we should return it
			if err != nil {
				return err
			}

			// If the verify function returns true, the CSR is valid and we can return
			if valid {
				return nil
			}

		}
	}
}
