package csrsigning

import (
	"context"
	"crypto/rsa"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	certv1 "k8s.io/api/certificates/v1"
	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileAutoApprove(ctx context.Context, log log.Logger, csr *certv1.CertificateSigningRequest,
	priv *rsa.PrivateKey, client client.Client,
) (ctrl.Result, error) {
	var result certv1.RequestConditionType

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
	if err := client.SubResource("approval").Update(ctx, csr); err != nil {
		log.Error(err, "Failed to update CSR approval status")
		return ctrl.Result{}, err
	}
	log.Info("Updated CSR approval status")
	return ctrl.Result{}, nil
}
