package controllers

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DNSRebalancerController struct {
	snap      snap.Snap
	waitReady func()
}

func NewDNSRebalancerController(snap snap.Snap, waitReady func()) *DNSRebalancerController {
	return &DNSRebalancerController{
		snap:      snap,
		waitReady: waitReady,
	}
}

func (c *DNSRebalancerController) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithValues("controller", "dns-rebalancer")
	log.Info("DNS rebalancer controller started")
	c.waitReady()

	k8sClient, err := c.snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Check if minimum 2 nodes are Ready
	log.Info("Waiting for at least 2 control-plane nodes to be ready")
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		readyCount, err := c.getNodesReadyCount(ctx, k8sClient)
		if err != nil {
			log.V(1).Info("Failed to get control plane counts while waiting", "error", err)
			return false, nil
		}
		log.V(1).Info("Checking node readiness", "readyNodes", readyCount)
		return readyCount >= 2, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for control plane nodes to be ready: %w", err)
	}
	log.Info("Control-plane nodes ready, checking CoreDNS distribution")

	needsRebalancing, err := c.coreDNSNeedsRebalancing(ctx, k8sClient)
	if err != nil {
		return fmt.Errorf("failed to check CoreDNS pods distribution: %w", err)
	}
	if !needsRebalancing {
		log.Info("CoreDNS pods are already balanced across control plane nodes")
		return nil
	}

	log.Info("Triggering CoreDNS deployment rollout restart to rebalance pods across control plane nodes")
	if err := k8sClient.RestartDeployment(ctx, "coredns", "kube-system"); err != nil {
		return fmt.Errorf("failed to restart CoreDNS deployment: %w", err)
	}

	return nil
}

func (c *DNSRebalancerController) coreDNSNeedsRebalancing(ctx context.Context, k8sClient *kubernetes.Client) (bool, error) {
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

func (c *DNSRebalancerController) getNodesReadyCount(ctx context.Context, k8sClient *kubernetes.Client) (readyNodeCount int, err error) {
	nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	readyNodeCount = 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				readyNodeCount++
				break
			}
		}
	}
	return readyNodeCount, nil
}
