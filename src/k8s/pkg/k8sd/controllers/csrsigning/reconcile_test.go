package csrsigning

import (
	"context"
	"crypto/rsa"
	"crypto/x509/pkix"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	certv1 "k8s.io/api/certificates/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8smock "github.com/canonical/k8s/pkg/client/k8s/mock"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
)

func TestCSRNotFound(t *testing.T) {
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{},
		&apierrors.StatusError{
			ErrStatus: v1.Status{
				Reason: v1.StatusReasonNotFound,
			},
		},
	)

	reconciler := &csrSigningReconciler{
		Client: k8sM,
	}

	g := NewWithT(t)

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestFailedToGetCSR(t *testing.T) {
	getErr := &apierrors.StatusError{
		ErrStatus: v1.Status{
			Reason: v1.StatusReasonInternalError,
		},
	}
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{},
		getErr,
	)

	reconciler := &csrSigningReconciler{
		Client: k8sM,
	}

	g := NewWithT(t)

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).To(MatchError(getErr))
}

func TestHasSignedCertificate(t *testing.T) {
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{
			Status: certv1.CertificateSigningRequestStatus{
				Certificate: []byte("cert"),
			},
		},
		nil,
	)

	reconciler := &csrSigningReconciler{
		Client: k8sM,
	}

	g := NewWithT(t)

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestSkipUnmanagedSignerName(t *testing.T) {
	unmanagedSignerName := "unknown"
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{
			Spec: certv1.CertificateSigningRequestSpec{
				SignerName: unmanagedSignerName,
			},
		},
		nil,
	)

	managedSigners := map[string]struct{}{
		"signer1": {},
		"signer2": {},
	}

	g := NewWithT(t)
	// just to make sure the test is correct
	g.Expect(managedSigners).ToNot(HaveKey(unmanagedSignerName))

	reconciler := &csrSigningReconciler{
		Client:             k8sM,
		managedSignerNames: managedSigners,
	}

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCertificateDenied(t *testing.T) {
	managedSigner := "managed-signer"
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{
			Spec: certv1.CertificateSigningRequestSpec{
				SignerName: managedSigner,
			},
			Status: certv1.CertificateSigningRequestStatus{
				Conditions: []certv1.CertificateSigningRequestCondition{
					{
						Type: certv1.CertificateDenied,
					},
				},
			},
		},
		nil,
	)

	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
	}

	g := NewWithT(t)

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCertificateFailed(t *testing.T) {
	managedSigner := "managed-signer"
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{
			Spec: certv1.CertificateSigningRequestSpec{
				SignerName: "k8sd.io/kubelet-serving",
			},
			Status: certv1.CertificateSigningRequestStatus{
				Conditions: []certv1.CertificateSigningRequestCondition{
					{
						Type: certv1.CertificateFailed,
					},
				},
			},
		},
		nil,
	)

	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
	}

	g := NewWithT(t)

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
}

func TestFailedToGetClusterConfig(t *testing.T) {
	managedSigner := "managed-signer"
	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		certv1.CertificateSigningRequest{
			Spec: certv1.CertificateSigningRequestSpec{
				SignerName: managedSigner,
			},
			Status: certv1.CertificateSigningRequestStatus{
				Conditions: []certv1.CertificateSigningRequestCondition{
					{
						Type: certv1.CertificateApproved,
					},
				},
			},
		},
		nil,
	)

	getCCErr := errors.New("failed to get cluster config")

	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
		getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
			return types.ClusterConfig{}, getCCErr
		},
	}

	g := NewWithT(t)

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).To(MatchError(getCCErr))
}

func TestNotApprovedCSR(t *testing.T) {
	t.Run("NoAutoApprove", func(t *testing.T) {
		managedSigner := "managed-signer"
		k8sM := k8smock.New(
			t,
			k8smock.NewSubResourceClientMock(nil),
			certv1.CertificateSigningRequest{
				Spec: certv1.CertificateSigningRequestSpec{
					SignerName: managedSigner,
				},
			},
			nil,
		)

		reconciler := &csrSigningReconciler{
			Client: k8sM,
			managedSignerNames: map[string]struct{}{
				managedSigner: {},
			},
			getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
				return types.ClusterConfig{}, nil
			},
		}

		g := NewWithT(t)

		result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

		g.Expect(result).To(Equal(ctrl.Result{RequeueAfter: requeueAfterWaitingForApproved}))
		g.Expect(err).ToNot(HaveOccurred())
	})

	t.Run("WithAutoApprove", func(t *testing.T) {
		managedSigner := "managed-signer"
		k8sM := k8smock.New(
			t,
			k8smock.NewSubResourceClientMock(nil),
			certv1.CertificateSigningRequest{
				Spec: certv1.CertificateSigningRequestSpec{
					SignerName: managedSigner,
				},
			},
			nil,
		)

		priv, _, err := pkiutil.GenerateRSAKey(2048)

		g := NewWithT(t)
		g.Expect(err).ToNot(HaveOccurred())

		var called bool
		reconciler := &csrSigningReconciler{
			Client: k8sM,
			managedSignerNames: map[string]struct{}{
				managedSigner: {},
			},
			getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
				return types.ClusterConfig{
					Annotations: map[string]string{
						"k8sd/v1alpha1/csrsigning/auto-approve": "true",
					},
					Certificates: types.Certificates{
						K8sdPrivateKey: ptr.To(priv),
					},
				}, nil
			},
			reconcileAutoApprove: func(ctx context.Context, l log.Logger, csr *certv1.CertificateSigningRequest, pk *rsa.PrivateKey, c client.Client, f func(*certv1.CertificateSigningRequest, *rsa.PrivateKey) error) (ctrl.Result, error) {
				called = true
				return ctrl.Result{}, nil
			},
		}

		_, _ = reconciler.Reconcile(context.Background(), getDefaultRequest())

		g.Expect(called).To(BeTrue())
	})
}

func TestInvalidCSR(t *testing.T) {
	managedSigner := "managed-signer"
	csr := certv1.CertificateSigningRequest{
		Spec: certv1.CertificateSigningRequestSpec{
			SignerName: managedSigner,
			Request:    []byte("invalid-csr"),
		},
		Status: certv1.CertificateSigningRequestStatus{
			Conditions: []certv1.CertificateSigningRequestCondition{
				{
					Type: certv1.CertificateApproved,
				},
			},
		},
	}

	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		csr,
		nil,
	)
	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
		getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
			return types.ClusterConfig{
				Certificates: types.Certificates{},
			}, nil
		},
	}

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g := NewWithT(t)
	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).To(HaveOccurred())
}

func TestInvalidCACertificate(t *testing.T) {
	csrPEM, _, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   "system:node:valid-node",
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		nil,
	)

	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred())

	managedSigner := "k8sd.io/kubelet-serving"
	csr := certv1.CertificateSigningRequest{
		Spec: certv1.CertificateSigningRequestSpec{
			SignerName: managedSigner,
			Request:    []byte(csrPEM),
		},
		Status: certv1.CertificateSigningRequestStatus{
			Conditions: []certv1.CertificateSigningRequestCondition{
				{
					Type: certv1.CertificateApproved,
				},
			},
		},
	}

	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		csr,
		nil,
	)
	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
		getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
			return types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: ptr.To("invalid"),
					CAKey:  ptr.To("invalid"),
				},
			}, nil
		},
	}

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).To(HaveOccurred())
}

func TestUpdateCSRFailed(t *testing.T) {
	csrPEM, _, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   "system:node:valid-node",
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		nil,
	)

	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred())

	managedSigner := "k8sd.io/kubelet-serving"
	csr := certv1.CertificateSigningRequest{
		Spec: certv1.CertificateSigningRequestSpec{
			SignerName: managedSigner,
			Request:    []byte(csrPEM),
		},
		Status: certv1.CertificateSigningRequestStatus{
			Conditions: []certv1.CertificateSigningRequestCondition{
				{
					Type: certv1.CertificateApproved,
				},
			},
		},
	}
	updateErr := errors.New("failed to update")

	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(
			updateErr,
		),
		csr,
		nil,
	)

	caCert, caKey, err := pkiutil.GenerateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca"}, 10, 2048)
	g.Expect(err).ToNot(HaveOccurred())

	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
		getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
			return types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: ptr.To(caCert),
					CAKey:  ptr.To(caKey),
				},
			}, nil
		},
	}

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).To(MatchError(updateErr))
	k8sM.AssertUpdateCalled(t)
}

func TestUpdateCSRSucceed(t *testing.T) {
	csrPEM, _, err := pkiutil.GenerateCSR(
		pkix.Name{
			CommonName:   "system:node:valid-node",
			Organization: []string{"system:nodes"},
		},
		2048,
		nil,
		nil,
	)

	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred())

	managedSigner := "k8sd.io/kubelet-serving"
	csr := certv1.CertificateSigningRequest{
		Spec: certv1.CertificateSigningRequestSpec{
			SignerName: managedSigner,
			Request:    []byte(csrPEM),
		},
		Status: certv1.CertificateSigningRequestStatus{
			Conditions: []certv1.CertificateSigningRequestCondition{
				{
					Type: certv1.CertificateApproved,
				},
			},
		},
	}

	k8sM := k8smock.New(
		t,
		k8smock.NewSubResourceClientMock(nil),
		csr,
		nil,
	)

	caCert, caKey, err := pkiutil.GenerateSelfSignedCA(pkix.Name{CommonName: "kubernetes-ca"}, 10, 2048)
	g.Expect(err).ToNot(HaveOccurred())

	reconciler := &csrSigningReconciler{
		Client: k8sM,
		managedSignerNames: map[string]struct{}{
			managedSigner: {},
		},
		getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
			return types.ClusterConfig{
				Certificates: types.Certificates{
					CACert: ptr.To(caCert),
					CAKey:  ptr.To(caKey),
				},
			}, nil
		},
	}

	result, err := reconciler.Reconcile(context.Background(), getDefaultRequest())

	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(err).ToNot(HaveOccurred())
	k8sM.AssertUpdateCalled(t)
}

func getDefaultRequest() ctrl.Request {
	return ctrl.Request{
		NamespacedName: k8stypes.NamespacedName{
			Name:      "csr-1",
			Namespace: "default",
		},
	}
}
