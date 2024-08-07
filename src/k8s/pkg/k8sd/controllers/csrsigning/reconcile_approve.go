package csrsigning

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	certv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *csrSigningReconciler) reconcileAutoApprove(ctx context.Context, log log.Logger, csr *certv1.CertificateSigningRequest, clusterConfig types.ClusterConfig) (ctrl.Result, error) {
	var result certv1.RequestConditionType

	keyPEM := clusterConfig.Certificates.GetK8sdPrivateKey()

	if keyPEM == "" {
		return ctrl.Result{}, fmt.Errorf("cluster RSA key not set")
	}

	priv, err := pkiutil.LoadRSAPrivateKey(keyPEM)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to load cluster RSA key: %w", err)
	}

	if err := validateCSR(csr, priv); err != nil {
		log.Error(err, "CSR is not valid")

		result = certv1.CertificateDenied
		csr.Status.Conditions = append(csr.Status.Conditions,
			certv1.CertificateSigningRequestCondition{
				Type:    certv1.CertificateDenied,
				Status:  v1.ConditionTrue,
				Reason:  "K8sdDeny",
				Message: fmt.Sprintf("CSR is not valid: %v", err.Error()),
			},
		)
	} else {
		result = certv1.CertificateApproved
		csr.Status.Conditions = append(csr.Status.Conditions,
			certv1.CertificateSigningRequestCondition{
				Type:    certv1.CertificateApproved,
				Status:  v1.ConditionTrue,
				Reason:  "K8sdApprove",
				Message: "CSR approved by k8sd",
			},
		)
	}

	log = log.WithValues("result", result)
	if err := r.Client.SubResource("approval").Update(ctx, csr); err != nil {
		log.Error(err, "Failed to update CSR approval status")
		return ctrl.Result{}, err
	}
	log.Info("Updated CSR approval status")
	return ctrl.Result{}, nil
}
