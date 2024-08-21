package csrsigning

import (
	"context"
	"crypto/rsa"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	certv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	k8smock "github.com/canonical/k8s/pkg/client/k8s/mock"
	"github.com/canonical/k8s/pkg/log"
)

func TestAutoApprove(t *testing.T) {
	for _, tc := range []struct {
		name        string
		csr         certv1.CertificateSigningRequest
		validateCSR func(obj *certv1.CertificateSigningRequest, priv *rsa.PrivateKey) error
		updateErr   error

		expCtrl      ctrl.Result
		expErr       error
		expCondition certv1.CertificateSigningRequestCondition
	}{
		{
			name: "InvalidCSR--UpdateSuccessful",
			csr:  certv1.CertificateSigningRequest{},
			validateCSR: func(obj *certv1.CertificateSigningRequest, priv *rsa.PrivateKey) error {
				return errors.New("invalid")
			},
			updateErr: nil,
			expCtrl:   ctrl.Result{},
			expErr:    nil,
			expCondition: certv1.CertificateSigningRequestCondition{
				Type:    certv1.CertificateDenied,
				Status:  v1.ConditionTrue,
				Reason:  "K8sdDeny",
				Message: "CSR is not valid: invalid",
			},
		},
		{
			name: "InvalidCSR--UpdateFailed",
			csr:  certv1.CertificateSigningRequest{},
			validateCSR: func(obj *certv1.CertificateSigningRequest, priv *rsa.PrivateKey) error {
				return errors.New("invalid")
			},
			updateErr: errors.New("failed to update"),
			expCtrl:   ctrl.Result{},
			expErr:    errors.New("failed to update"),
			expCondition: certv1.CertificateSigningRequestCondition{
				Type:    certv1.CertificateDenied,
				Status:  v1.ConditionTrue,
				Reason:  "K8sdDeny",
				Message: "CSR is not valid: invalid",
			},
		},
		{
			name: "ValidCSR--UpdateSuccessful",
			csr:  certv1.CertificateSigningRequest{},
			validateCSR: func(obj *certv1.CertificateSigningRequest, priv *rsa.PrivateKey) error {
				return nil
			},
			updateErr: nil,
			expCtrl:   ctrl.Result{},
			expErr:    nil,
			expCondition: certv1.CertificateSigningRequestCondition{
				Type:    certv1.CertificateApproved,
				Status:  v1.ConditionTrue,
				Reason:  "K8sdApprove",
				Message: "CSR approved by k8sd",
			},
		},
		{
			name: "ValidCSR--UpdateFailed",
			csr:  certv1.CertificateSigningRequest{},
			validateCSR: func(obj *certv1.CertificateSigningRequest, priv *rsa.PrivateKey) error {
				return nil
			},
			updateErr: errors.New("failed to update"),
			expCtrl:   ctrl.Result{},
			expErr:    errors.New("failed to update"),
			expCondition: certv1.CertificateSigningRequestCondition{
				Type:    certv1.CertificateApproved,
				Status:  v1.ConditionTrue,
				Reason:  "K8sdApprove",
				Message: "CSR approved by k8sd",
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
				nil, // not important, validateCSR is mocked
				k8sM,
				tc.validateCSR,
			)

			g := NewWithT(t)

			k8sM.AssertUpdateCalled(t)
			g.Expect(result).To(Equal(tc.expCtrl))
			if tc.expErr == nil {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err).To(MatchError(tc.expErr))
			}
			g.Expect(tc.csr.Status.Conditions).To(ContainElement(tc.expCondition))
		})
	}
}
