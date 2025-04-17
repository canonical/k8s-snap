package controllers

import (
	"context"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/snap"
)

func getNewK8sClientWithRetries(ctx context.Context, snapObj snap.Snap, admin bool) (*kubernetes.Client, error) {
	for {
		var err error
		var client *kubernetes.Client
		if admin {
			// use admin client
			client, err = snapObj.KubernetesClient("kube-system")
		} else {
			// use node client
			client, err = snapObj.KubernetesNodeClient("kube-system")
		}

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
