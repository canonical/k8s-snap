package csrsigning

import (
	"context"
	"crypto/rsa"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/go-logr/logr"
	certv1 "k8s.io/api/certificates/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type csrSigningReconciler struct {
	Manager            manager.Manager
	Logger             logr.Logger
	Client             client.Client
	managedSignerNames map[string]struct{}

	getClusterConfig     func(context.Context) (types.ClusterConfig, error)
	reconcileAutoApprove func(context.Context, log.Logger, *certv1.CertificateSigningRequest, *rsa.PrivateKey, client.Client) (ctrl.Result, error)
}

var managedSignerNames = map[string]struct{}{
	"k8sd.io/kubelet-serving":   {},
	"k8sd.io/kubelet-client":    {},
	"k8sd.io/kube-proxy-client": {},
}

func (r *csrSigningReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certv1.CertificateSigningRequest{}).
		WithEventFilter(predicate.NewPredicateFuncs(func(object client.Object) bool {
			if csr, ok := object.(*certv1.CertificateSigningRequest); !ok {
				return false
			} else if _, ok := r.managedSignerNames[csr.Spec.SignerName]; !ok {
				return false
			}
			return true
		})).
		Complete(r)
}
