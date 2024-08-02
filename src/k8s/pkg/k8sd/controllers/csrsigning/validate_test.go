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
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	csrPEM, _, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   "system:node:valid-node",
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("failed to generate test CSR: %v", err)
	}

	tests := []struct {
		name       string
		csr        *certv1.CertificateSigningRequest
		wantErr    bool
		errMessage string
	}{
		{
			name: "Valid CSR",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(t, &key.PublicKey, csrPEM),
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
			wantErr: false,
		},
		{
			name: "Bad encrypted signature",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(t, &key.PublicKey, "changedCSR"),
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
			wantErr:    true,
			errMessage: "CSR signature does not match",
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
			wantErr:    true,
			errMessage: "failed to decrypt signature",
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
			wantErr:    true,
			errMessage: "failed to decrypt signature",
		},
		{
			name: "Missing k8sd.io/node annotation",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(t, &key.PublicKey, csrPEM),
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
			wantErr:    true,
			errMessage: "k8sd.io/node annotation missing from CSR object",
		},
		{
			name: "Invalid node name in CSR",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/node":      "invalid-node!",
						"k8sd.io/signature": mustCreateEncryptedSignature(t, &key.PublicKey, csrPEM),
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
			wantErr:    true,
			errMessage: "CSR has invalid node name",
		},
		{
			name: "Invalid Signer Name",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(t, &key.PublicKey, csrPEM),
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
			wantErr:    true,
			errMessage: "CSR has unknown signerName",
		},
		{
			name: "Invalid Usages",
			csr: &certv1.CertificateSigningRequest{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"k8sd.io/signature": mustCreateEncryptedSignature(t, &key.PublicKey, csrPEM),
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
			wantErr:    true,
			errMessage: "CSR usages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			err := validateCSR(tt.csr, key)
			if tt.wantErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.errMessage))
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func mustCreateEncryptedSignature(t *testing.T, pub *rsa.PublicKey, csrPEM string) string {
	// calculate sha256 sum of CSR request
	hash := sha256.New()
	if _, err := hash.Write([]byte(csrPEM)); err != nil {
		t.Fatalf("failed to compute sha256: %v", err)
	}

	// encrypt the hash with the public cluster RSA key
	signature, err := rsa.EncryptPKCS1v15(rand.Reader, pub, hash.Sum(nil))
	if err != nil {
		t.Fatalf("failed to encrypt csr signature: %v", err)
	}

	return string(signature)
}
