package api

import (
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/v3/state"
	"golang.org/x/sync/errgroup"
	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// postApproveWorkerCSR approves the worker node CSR for the specified seed.
// The certificate approval process follows these steps:
// 1. The CAPI provider calls the /x/capi/refresh-certs/plan endpoint from a
// worker node, which generates a CSR and creates a CertificateSigningRequest
// object in the cluster.
// 2. The CAPI provider then calls the /k8sd/refresh-certs/run endpoint with
// the seed. This endpoint waits until the CSR is approved and the certificate
// is signed. Note that this is a blocking call.
// 3. The CAPI provider calls the /x/capi/refresh-certs/approve endpoint from
// any control plane node to approve the CSR.
// 4. The /x/capi/refresh-certs/run endpoint completes and returns once the
// certificate is approved and signed.
func (e *Endpoints) postApproveWorkerCSR(s state.State, r *http.Request) response.Response {
	snap := e.provider.Snap()

	req := apiv1.ClusterAPIApproveWorkerCSRRequest{}

	if err := utils.NewStrictJSONDecoder(r.Body).Decode(&req); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	if err := r.Body.Close(); err != nil {
		return response.InternalError(fmt.Errorf("failed to close request body: %w", err))
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
			if err := client.WatchCertificateSigningRequest(
				ctx,
				csrName,
				func(request *certv1.CertificateSigningRequest) (bool, error) {
					request.Status.Conditions = append(request.Status.Conditions, certv1.CertificateSigningRequestCondition{
						Type:           certv1.CertificateApproved,
						Status:         corev1.ConditionTrue,
						Reason:         "ApprovedByCK8sCAPI",
						Message:        "This CSR was approved by the Canonical Kubernetes CAPI Provider",
						LastUpdateTime: metav1.Now(),
					})
					_, err := client.CertificatesV1().CertificateSigningRequests().UpdateApproval(ctx, csrName, request, metav1.UpdateOptions{})
					if err != nil {
						return false, fmt.Errorf("failed to update CSR %s: %w", csrName, err)
					}
					return true, nil
				},
			); err != nil {
				return fmt.Errorf("certificate signing request failed: %w", err)
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return response.InternalError(fmt.Errorf("failed to approve worker node CSR: %w", err))
	}

	return response.SyncResponse(true, apiv1.ClusterAPIApproveWorkerCSRResponse{})
}
