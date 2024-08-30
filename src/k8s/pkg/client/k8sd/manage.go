package k8sd

import (
	"context"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
)

func (c *k8sd) RefreshCertificatesPlan(ctx context.Context, request apiv1.RefreshCertificatesPlanRequest) (apiv1.RefreshCertificatesPlanResponse, error) {
	return query(ctx, c, "POST", apiv1.RefreshCertificatesPlanRPC, request, &apiv1.RefreshCertificatesPlanResponse{})
}

func (c *k8sd) RefreshCertificatesRun(ctx context.Context, request apiv1.RefreshCertificatesRunRequest) (apiv1.RefreshCertificatesRunResponse, error) {
	return query(ctx, c, "POST", apiv1.RefreshCertificatesRunRPC, request, &apiv1.RefreshCertificatesRunResponse{})
}
