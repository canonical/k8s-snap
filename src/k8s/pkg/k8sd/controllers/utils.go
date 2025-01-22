package controllers

import (
	"context"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/snap"
)

func getNewK8sClientWithRetries(ctx context.Context, snapObj snap.Snap) (*kubernetes.Client, error) {
	for {
		client, err := snapObj.KubernetesNodeClient("kube-system")
		if err == nil {
			return client, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}
