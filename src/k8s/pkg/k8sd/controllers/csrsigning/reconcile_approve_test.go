package csrsigning

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509/pkix"
	"errors"
	"testing"

	k8smock "github.com/canonical/k8s/pkg/k8sd/controllers/csrsigning/test"
	"github.com/canonical/k8s/pkg/log"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	. "github.com/onsi/gomega"
	certv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestAutoApprove(t *testing.T) {
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

	for _, tc := range []struct {
		name      string
		csr       certv1.CertificateSigningRequest
		updateErr error

		expectResult    ctrl.Result
		expectErr       error
		expectCondition certv1.CertificateSigningRequestCondition
	}{
		{
			name: "InvalidCSR/UpdateSuccessful",
			csr:  certv1.CertificateSigningRequest{}, // invalid csr

			expectResult: ctrl.Result{},
			expectCondition: certv1.CertificateSigningRequestCondition{
				Type:   certv1.CertificateDenied,
				Status: v1.ConditionTrue,
				Reason: "K8sdDeny",
			},
		},
		{
			name: "InvalidCSR/UpdateFailed",
			csr:  certv1.CertificateSigningRequest{}, // invalid csr

			updateErr:    errors.New("failed to update"),
			expectResult: ctrl.Result{},
			expectErr:    errors.New("failed to update"),
			expectCondition: certv1.CertificateSigningRequestCondition{
				Type:   certv1.CertificateDenied,
				Status: v1.ConditionTrue,
				Reason: "K8sdDeny",
			},
		},
		{
			name: "ValidCSR/UpdateSuccessful",
			csr: certv1.CertificateSigningRequest{
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

			expectResult: ctrl.Result{},
			expectCondition: certv1.CertificateSigningRequestCondition{
				Type:   certv1.CertificateApproved,
				Status: v1.ConditionTrue,
				Reason: "K8sdApprove",
			},
		},
		{
			name: "ValidCSR/UpdateFailed",
			csr: certv1.CertificateSigningRequest{
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

			updateErr:    errors.New("failed to update"),
			expectResult: ctrl.Result{},
			expectErr:    errors.New("failed to update"),
			expectCondition: certv1.CertificateSigningRequestCondition{
				Type:   certv1.CertificateApproved,
				Status: v1.ConditionTrue,
				Reason: "K8sdApprove",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			k8sM := k8smock.New(
				t,
				k8smock.NewSubResourceClientMock(tc.updateErr),
				tc.csr,
				nil, // we don't call get in reconcileAutoApprove
			)

			result, err := reconcileAutoApprove(
				context.Background(),
				log.L(),
				&tc.csr,
				key,
				k8sM,
			)

			g := NewWithT(t)
			k8sM.AssertUpdateCalled(t)
			g.Expect(result).To(Equal(tc.expectResult))
			if tc.expectErr == nil {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(MatchError(tc.expectErr))
			}
			g.Expect(containsCondition(tc.csr.Status.Conditions, tc.expectCondition)).To(BeTrue(), "expected condition not found")
		})
	}
}

func containsCondition(cc []certv1.CertificateSigningRequestCondition, c certv1.CertificateSigningRequestCondition) bool {
	for _, cond := range cc {
		if cond.Type == c.Type &&
			cond.Status == c.Status &&
			cond.Reason == c.Reason {
			return true
		}
	}
	return false
}
