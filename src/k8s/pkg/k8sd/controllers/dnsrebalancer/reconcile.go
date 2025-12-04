package dnsrebalancer

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/log"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile implements the reconciliation loop.
func (r *Controller) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("node", req.Name)

	// Count ready nodes
	nodeList := &corev1.NodeList{}
	if err := r.client.List(ctx, nodeList); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to list nodes")
		return ctrl.Result{}, err
	}

	readyCount := countReadyNodes(nodeList)
	if readyCount < 2 {
		log.V(1).Info("Less than 2 nodes ready, skipping rebalancing check", "readyCount", readyCount)
		return ctrl.Result{}, nil
	}

	log.V(1).Info("Sufficient nodes ready, checking CoreDNS distribution", "readyCount", readyCount)

	// Check if rebalancing is needed
	needsRebalancing, err := r.coreDNSNeedsRebalancing(ctx)
	if err != nil {
		log.Error(err, "Failed to check CoreDNS pods distribution")
		return ctrl.Result{}, err
	}

	if !needsRebalancing {
		log.V(1).Info("CoreDNS pods are already balanced across nodes")
		return ctrl.Result{}, nil
	}

	log.Info("CoreDNS pods need rebalancing, triggering deployment rollout restart")

	// Get kubernetes client to restart deployment
	k8sClient, err := r.snap.KubernetesClient("")
	if err != nil {
		log.Error(err, "Failed to get Kubernetes client")
		return ctrl.Result{}, err
	}

	if err := k8sClient.RestartDeployment(ctx, "coredns", "kube-system"); err != nil {
		log.Error(err, "Failed to restart CoreDNS deployment, will retry on next reconciliation")
		return ctrl.Result{}, err
	}

	log.Info("Successfully triggered CoreDNS deployment restart")
	return ctrl.Result{}, nil
}

// coreDNSNeedsRebalancing checks if CoreDNS pods need rebalancing.
func (r *Controller) coreDNSNeedsRebalancing(ctx context.Context) (bool, error) {
	pods := &corev1.PodList{}
	if err := r.client.List(ctx, pods, client.InNamespace("kube-system"), client.MatchingLabels{"k8s-app": "coredns"}); err != nil {
		return false, fmt.Errorf("failed to list CoreDNS pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return false, fmt.Errorf("no CoreDNS pods found")
	}

	// Filter scheduled pods (NodeName != ""). Pending pods can cause false positives.
	scheduled := make([]corev1.Pod, 0, len(pods.Items))
	for _, pod := range pods.Items {
		if pod.Spec.NodeName != "" {
			scheduled = append(scheduled, pod)
		}
	}

	// If fewer than 2 pods are scheduled, we cannot assess imbalance yet.
	if len(scheduled) < 2 {
		return false, nil
	}

	// Check if all scheduled pods are on the same node
	firstNodeName := scheduled[0].Spec.NodeName
	for _, pod := range scheduled[1:] {
		if pod.Spec.NodeName != firstNodeName {
			// Pods are on different nodes - no rebalancing needed
			return false, nil
		}
	}
	// All scheduled pods are on the same node - rebalancing needed
	return true, nil
}

// countReadyNodes counts the number of nodes with Ready condition.
func countReadyNodes(nodeList *corev1.NodeList) int {
	readyCount := 0
	for _, node := range nodeList.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyCount++
				break
			}
		}
	}
	return readyCount
}
