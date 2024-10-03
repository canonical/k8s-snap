package api

import (
	"fmt"
	"net/http"

	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
	"golang.org/x/sync/errgroup"
	certv1 "k8s.io/api/certificates/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ApproveWorkerCSRRequest struct {
	Seed int
}

type ApproveWorkerCSRResponse struct{}

func (e *Endpoints) postApproveWorkerCSR(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := ApproveWorkerCSRRequest{}
	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	client, err := snap.KubernetesClient("")
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to get Kubernetes client: %w", err))
	}

	g, ctx := errgroup.WithContext(r.Context())

	// CSR names
	csrNames := []string{
		fmt.Sprintf("k8sd-%d-worker-kubelet-serving", req.Seed),
		fmt.Sprintf("k8sd-%d-worker-kubelet-client", req.Seed),
		fmt.Sprintf("k8sd-%d-worker-kube-proxy-client", req.Seed),
	}

	for _, csrName := range csrNames {
		csrName := csrName
		g.Go(func() error {
			csrObject, err := client.CertificatesV1().CertificateSigningRequests().Get(ctx, csrName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get CSR %s: %w", csrObject, err)
			}
			// Approve the CSR
			for _, condition := range csrObject.Status.Conditions {
				if condition.Type == certv1.CertificateApproved {
					return fmt.Errorf("CSR %s already approved", csrName)
				}
			}

			csrObject.Status.Conditions = append(csrObject.Status.Conditions, certv1.CertificateSigningRequestCondition{
				Type:           certv1.CertificateApproved,
				Status:         "True",
				Reason:         "ApprovedByCK8sCAPI",
				Message:        "This CSR was approved by Canonical Kubernetes CAPI Provider",
				LastUpdateTime: metav1.Now(),
			})

			_, err = client.CertificatesV1().CertificateSigningRequests().UpdateApproval(r.Context(), csrName, csrObject, metav1.UpdateOptions{})

			return nil

		})

	}

	if err := g.Wait(); err != nil {
		return response.InternalError(fmt.Errorf("failed to approve worker node CSR: %w", err))
	}

	return response.SyncResponse(true, ApproveWorkerCSRResponse{})
}
