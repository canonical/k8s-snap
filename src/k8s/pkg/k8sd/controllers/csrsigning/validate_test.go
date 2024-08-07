package csrsigning

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509/pkix"
	"testing"

	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	. "github.com/onsi/gomega"
	certv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateCSREncryption(t *testing.T) {
	g := NewWithT(t)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	g.Expect(err).NotTo(HaveOccurred())

	csrPEM, _, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   "system:node:valid-node",
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		nil,
	)
	g.Expect(err).NotTo(HaveOccurred())

	tests := []struct {
		name               string
		csr                *certv1.CertificateSigningRequest
		expectErr          bool
		expectedErrMessage string
	}{
		{
			name: "Valid CSR",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(g, &key.PublicKey, csrPEM),
						"k8sd.io/node":      "valid-node",
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr: false,
		},
		{
			name: "Bad encrypted signature",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(g, &key.PublicKey, "changedCSR"),
						"k8sd.io/node":      "valid-node",
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr:          true,
			expectedErrMessage: "CSR signature does not match",
		},
		{
			name: "Invalid Signature",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": "invalid-signature",
						"k8sd.io/node":      "valid-node",
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr:          true,
			expectedErrMessage: "failed to decrypt signature",
		},
		{
			name: "Missing Signature",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/node": "valid-node",
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr:          true,
			expectedErrMessage: "failed to decrypt signature",
		},
		{
			name: "Missing k8sd.io/node annotation",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(g, &key.PublicKey, csrPEM),
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr:          true,
			expectedErrMessage: "k8sd.io/node annotation missing from CSR object",
		},
		{
			name: "Invalid node name in CSR",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/node":      "invalid-node!",
						"k8sd.io/signature": mustCreateEncryptedSignature(g, &key.PublicKey, csrPEM),
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:invalid-node!",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr:          true,
			expectedErrMessage: "CSR has invalid node name",
		},
		{
			name: "Invalid Signer Name",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(g, &key.PublicKey, csrPEM),
						"k8sd.io/node":      "valid-node",
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "invalid-signer-name",
					Usages:     []certv1.KeyUsage{certv1.UsageServerAuth, certv1.UsageDigitalSignature, certv1.UsageKeyEncipherment},
				},
			},
			expectErr:          true,
			expectedErrMessage: "CSR has unknown signerName",
		},
		{
			name: "Invalid Usages",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(g, &key.PublicKey, csrPEM),
						"k8sd.io/node":      "valid-node",
					},
				},
				Spec: certv1.CertificateSigningRequestSpec{
					Request:    []byte(csrPEM),
					Username:   "system:node:valid-node",
					Groups:     []string{"system:nodes"},
					SignerName: "k8sd.io/kubelet-serving",
					Usages:     []certv1.KeyUsage{certv1.UsageClientAuth}, // Invalid usages
				},
			},
			expectErr:          true,
			expectedErrMessage: "CSR usages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateCSR(tt.csr, key)
			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.expectedErrMessage))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func mustCreateEncryptedSignature(g Gomega, pub *rsa.PublicKey, csrPEM string) string {
	// calculate sha256 sum of CSR request
	hash := sha256.New()
	_, err := hash.Write([]byte(csrPEM))
	g.Expect(err).NotTo(HaveOccurred())

	// encrypt the hash with the public cluster RSA key
	signature, err := rsa.EncryptPKCS1v15(rand.Reader, pub, hash.Sum(nil))
	g.Expect(err).NotTo(HaveOccurred())

	return string(signature)
}
