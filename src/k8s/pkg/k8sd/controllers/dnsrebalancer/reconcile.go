package dnsrebalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const minRebalanceInterval = 30 * time.Second

// corednsPodLabels selects the CoreDNS pods managed by the ck-dns chart.
var corednsPodLabels = map[string]string{"k8s-app": coredns.DeploymentName, "app.kubernetes.io/instance": coredns.Chart.Name}

// rebalanceAttempt identifies our rollout and its first settled pod set.
type rebalanceAttempt struct {
	restartedAt   string
	deploymentUID types.UID
	generation    int64
	nodes         map[string]types.UID
	pods          map[types.UID]struct{}
}

// Reconcile implements the reconciliation loop.
func (r *controller) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	if r.getClusterConfig == nil {
		return ctrl.Result{}, nil
	}
	config, err := r.getClusterConfig(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get cluster config: %w", err)
	}
	if !config.DNS.GetEnabled() {
		r.rebalance = nil
		return ctrl.Result{}, nil
	}

	nodes := &corev1.NodeList{}
	if err := r.client.List(ctx, nodes); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list nodes: %w", err)
	}
	schedulableCount := 0
	nodesUnchanged := r.rebalance != nil
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if nodeSchedulableForCoreDNS(node) {
			schedulableCount++
			if r.rebalance != nil {
				uid, ok := r.rebalance.nodes[node.Name]
				if !ok || uid != node.UID {
					nodesUnchanged = false
				}
			}
		}
	}
	if r.rebalance != nil && (!nodesUnchanged || schedulableCount != len(r.rebalance.nodes)) {
		r.rebalance = nil
	}
	if schedulableCount < 2 {
		return ctrl.Result{}, nil
	}

	deployment := &appsv1.Deployment{}
	if err := r.client.Get(ctx, client.ObjectKey{Namespace: coredns.Chart.Namespace, Name: coredns.DeploymentName}, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			r.rebalance = nil
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get CoreDNS deployment: %w", err)
	}
	if r.rebalance != nil && deployment.UID == r.rebalance.deploymentUID && deployment.Generation < r.rebalance.generation {
		return ctrl.Result{}, nil
	}
	restartedAt := deployment.Spec.Template.Annotations[kubernetes.RestartedAtAnnotation]
	if r.rebalance != nil && (deployment.UID != r.rebalance.deploymentUID ||
		deployment.Generation != r.rebalance.generation || restartedAt != r.rebalance.restartedAt) {
		r.rebalance = nil
	}
	desiredReplicas := int32(1)
	if deployment.Spec.Replicas != nil {
		desiredReplicas = *deployment.Spec.Replicas
	}
	if desiredReplicas < 2 ||
		deployment.Status.ObservedGeneration < deployment.Generation ||
		deployment.Status.UpdatedReplicas != desiredReplicas ||
		deployment.Status.Replicas != desiredReplicas ||
		deployment.Status.ReadyReplicas != desiredReplicas ||
		deployment.Status.AvailableReplicas != desiredReplicas ||
		deployment.Status.UnavailableReplicas > 0 {
		return ctrl.Result{}, nil
	}

	pods := &corev1.PodList{}
	if err := r.client.List(ctx, pods, client.InNamespace(coredns.Chart.Namespace), client.MatchingLabels(corednsPodLabels)); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to list CoreDNS pods: %w", err)
	}
	now := r.now()
	var wait time.Duration
	if restarted, err := time.Parse(time.RFC3339, restartedAt); err == nil {
		wait = max(wait, restarted.Add(minRebalanceInterval).Sub(now))
	}
	activeCount := 0
	firstNode := ""
	coLocated := true
	ownRollout := r.rebalance != nil
	podsUnchanged := r.rebalance != nil && r.rebalance.pods != nil
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded {
			continue
		}
		activeCount++
		if pod.DeletionTimestamp != nil || pod.Spec.NodeName == "" || !kubernetes.IsPodReady(pod) {
			return ctrl.Result{}, nil
		}
		if !pod.CreationTimestamp.IsZero() {
			wait = max(wait, pod.CreationTimestamp.Add(minRebalanceInterval).Sub(now))
		}
		if firstNode == "" {
			firstNode = pod.Spec.NodeName
		} else if firstNode != pod.Spec.NodeName {
			coLocated = false
		}
		if r.rebalance != nil {
			ownRollout = ownRollout && pod.Annotations[kubernetes.RestartedAtAnnotation] == r.rebalance.restartedAt
			if _, ok := r.rebalance.pods[pod.UID]; !ok {
				podsUnchanged = false
			}
		}
	}
	if activeCount != int(desiredReplicas) {
		return ctrl.Result{}, nil
	}
	if wait > 0 {
		return ctrl.Result{RequeueAfter: wait}, nil
	}
	if !coLocated {
		r.rebalance = nil
		return ctrl.Result{}, nil
	}
	if r.rebalance != nil {
		// Cached pods may still precede our rollout; wait for its pod events.
		if !ownRollout {
			return ctrl.Result{}, nil
		}
		if r.rebalance.pods == nil {
			r.rebalance.pods = make(map[types.UID]struct{}, activeCount)
			for i := range pods.Items {
				pod := &pods.Items[i]
				if pod.Status.Phase != corev1.PodFailed && pod.Status.Phase != corev1.PodSucceeded {
					r.rebalance.pods[pod.UID] = struct{}{}
				}
			}
			return ctrl.Result{}, nil
		}
		if podsUnchanged && activeCount == len(r.rebalance.pods) {
			return ctrl.Result{}, nil
		}
	}

	log.Info("CoreDNS pods need rebalancing, triggering deployment rollout restart")
	kubernetes.SetDeploymentRestartedAt(deployment, now)
	if err := r.client.Update(ctx, deployment); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to restart CoreDNS deployment: %w", err)
	}
	r.rebalance = &rebalanceAttempt{
		restartedAt:   deployment.Spec.Template.Annotations[kubernetes.RestartedAtAnnotation],
		deploymentUID: deployment.UID,
		generation:    deployment.Generation,
		nodes:         make(map[string]types.UID, schedulableCount),
	}
	for i := range nodes.Items {
		node := &nodes.Items[i]
		if nodeSchedulableForCoreDNS(node) {
			r.rebalance.nodes[node.Name] = node.UID
		}
	}
	return ctrl.Result{}, nil
}

func nodeSchedulableForCoreDNS(node *corev1.Node) bool {
	if !kubernetes.IsNodeReady(node) || node.Spec.Unschedulable {
		return false
	}
	// This release's CoreDNS chart does not set any scheduling tolerations.
	for i := range node.Spec.Taints {
		effect := node.Spec.Taints[i].Effect
		if effect == corev1.TaintEffectNoSchedule || effect == corev1.TaintEffectNoExecute {
			return false
		}
	}
	return true
}
