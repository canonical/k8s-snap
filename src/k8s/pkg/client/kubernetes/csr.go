package kubernetes

import (
	"context"
	"fmt"

	certificatesv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) CreateCertificateSigningRequest(ctx context.Context, name string, csrPEM []byte, usages []certificatesv1.KeyUsage, groups []string, signerName string) (*certificatesv1.CertificateSigningRequest, error) {
	csr, err := c.CertificatesV1().CertificateSigningRequests().Create(ctx, &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:    csrPEM,
			Usages:     usages,
			Groups:     groups,
			SignerName: signerName,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate signing request %s: %w", name, err)
	}

	return csr, nil
}

func (c *Client) GetCertificateSigningRequest(ctx context.Context, name string) (*certificatesv1.CertificateSigningRequest, error) {
	csr, err := c.CertificatesV1().CertificateSigningRequests().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate signing request %s: %w", name, err)
	}

	return csr, nil
}
