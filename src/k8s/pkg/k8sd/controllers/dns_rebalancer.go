package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/control"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
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
	log.Info("Waiting for at least 2 nodes to be ready")
	if err := control.WaitUntilReady(ctx, func() (bool, error) {
		readyCount, err := c.getNodesReadyCount(ctx, k8sClient)
		if err != nil {
			log.V(1).Info("Failed to get node ready counts while waiting", "error", err)
			return false, nil
		}
		log.V(1).Info("Checking node readiness", "readyNodes", readyCount)
		return readyCount >= 2, nil
	}); err != nil {
		return fmt.Errorf("failed to wait for nodes to be ready: %w", err)
	}
	log.Info("At least two nodes ready, attempting leader election to ensure CoreDNS is balanced across nodes")

	// Acquire leader election lock first to ensure only one controller instance checks and restarts CoreDNS
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "coredns-rebalancer",
			Namespace: "kube-system",
		},
		Client: k8sClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: c.snap.Hostname(),
		},
	}

	leaderElector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          lock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Info("Acquired leader election, checking CoreDNS distribution")

				needsRebalancing, err := c.coreDNSNeedsRebalancing(ctx, k8sClient)
				if err != nil {
					log.Error(err, "Failed to check CoreDNS pods distribution")
					return
				}

				if !needsRebalancing {
					log.Info("CoreDNS pods are already balanced across nodes")
					return
				}

				log.Info("CoreDNS pods need rebalancing, triggering deployment rollout restart")
				if err := k8sClient.RestartDeployment(ctx, "coredns", "kube-system"); err != nil {
					log.Error(err, "Failed to restart CoreDNS deployment")
				} else {
					log.Info("Successfully triggered CoreDNS deployment restart")
				}
			},
			OnStoppedLeading: func() {
				log.Info("Lost leader election")
			},
		},
		ReleaseOnCancel: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create leader elector: %w", err)
	}

	// Run leader election (this will block until we become leader and execute the check/restart, or context is cancelled)
	leaderElector.Run(ctx)

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

	// Consider only scheduled pods (NodeName != ""). Pending pods can cause false positives.
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

func (c *DNSRebalancerController) getNodesReadyCount(ctx context.Context, k8sClient *kubernetes.Client) (readyNodeCount int, err error) {
	nodes, err := k8sClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	readyNodeCount = 0
	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				readyNodeCount++
				break
			}
		}
	}
	return readyNodeCount, nil
}
