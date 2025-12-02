package controllers

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DNSRebalancerController struct {
	snap snap.Snap
}

func NewDNSRebalancerController(snap snap.Snap) *DNSRebalancerController {
	return &DNSRebalancerController{
		snap: snap,
	}
}

func (c *DNSRebalancerController) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithValues("step", "coredns-rebalance")
	// Check if minimum 2 nodes are Ready
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		readyCount, err := c.controlPlaneReadyCount(ctx)
		if err != nil {
			log.V(1).Info("Failed to get control plane counts while waiting", "error", err)
			return false, nil
		}
		log.V(1).Info("Checking control plane readiness", "readyControlPlaneNodes", readyCount)
		return readyCount >= 2, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for control plane nodes to be ready: %w", err)
	}

	needsRebalancing, err := c.coreDNSNeedsRebalancing(ctx)
	if err != nil {
		return fmt.Errorf("failed to check CoreDNS pods distribution: %w", err)
	}
	if !needsRebalancing {
		log.V(1).Info("CoreDNS pods are already balanced across control plane nodes")
		return nil
	}

	log.Info("Triggering CoreDNS deployment rollout restart to rebalance pods across control plane nodes")
	k8sClient, err := c.snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := k8sClient.RestartDeployment(ctx, "coredns", "kube-system"); err != nil {
		return fmt.Errorf("failed to restart CoreDNS deployment: %w", err)
	}

	return nil
}

func (c *DNSRebalancerController) coreDNSNeedsRebalancing(ctx context.Context) (bool, error) {
	k8sClient, err := c.snap.KubernetesClient("")
	if err != nil {
		return false, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	pods, err := k8sClient.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app=coredns",
	})
	if err != nil {
		return false, fmt.Errorf("failed to list CoreDNS pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return false, fmt.Errorf("no CoreDNS pods found")
	}

	// Check if all pods are on the same node
	firstNodeName := pods.Items[0].Spec.NodeName
	for _, pod := range pods.Items[1:] {
		if pod.Spec.NodeName != firstNodeName {
			// Pods are on different nodes - no rebalancing needed
			return false, nil
		}
	}
	// All pods are on the same node - rebalancing needed
	return true, nil
}

func (c *DNSRebalancerController) controlPlaneReadyCount(ctx context.Context) (readyControlPlaneCount int, err error) {
	k8sClient, err := c.snap.KubernetesClient("")
	if err != nil {
		return 0, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	// Count only control plane nodes in Ready state
	readyControlPlaneCount = 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				readyControlPlaneCount++
				break
			}
		}
	}
	return readyControlPlaneCount, nil
}
