package coredns

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckDNS checks the CoreDNS deployment in the cluster.
func CheckDNS(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	for {
		pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		} else {
			for _, pod := range pods.Items {
				if strings.Contains(pod.Name, "coredns") {
					err := client.WaitForPodRunning(ctx, "kube-system", metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=coredns"})
					if err != nil {
						return fmt.Errorf("failed to wait for pod %q to be ready: %w", pod.Name, err)
					}
					return nil
				}
			}
		}

		time.Sleep(5 * time.Second)
	}

}
