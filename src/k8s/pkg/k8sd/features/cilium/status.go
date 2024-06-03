package cilium

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/canonical/k8s/pkg/snap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CheckNetwork(ctx context.Context, snap snap.Snap) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ciliumPods := map[string]string{
		"cilium-operator": "io.cilium/app=operator",
		"cilium":          "k8s-app=cilium",
	}

	for ciliumPod, selector := range ciliumPods {
		fmt.Printf("Checking for pod %q with selector %q\n", ciliumPod, selector)
		for {
			pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("failed to list pods: %w", err)
			} else {
				for _, pod := range pods.Items {
					if strings.Contains(pod.Name, ciliumPod) {
						fmt.Printf("Found pod %q. Checking if it's ready...\n", pod.Name)
						err := client.WaitForPodRunning(ctx, "kube-system", metav1.ListOptions{LabelSelector: selector})
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
	return nil
}
