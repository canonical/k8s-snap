package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/log"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
)

func (c *Client) GetNode(ctx context.Context, nodeName string) (*corev1.Node, error) {
	return c.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
}

// DeleteNode will remove a node from the kubernetes cluster.
// DeleteNode will retry if there is a conflict on the resource
// DeleteNode will retry if an internal server error occured (maximum of 5 times).
// DeleteNode will not fail if the node does not exist.
func (c *Client) DeleteNode(ctx context.Context, nodeName string) error {
	tries := 0
	retriable := func(err error) bool {
		if apierrors.IsConflict(err) {
			return true
		}

		tries++
		return apierrors.IsInternalError(err) && tries <= 5
	}

	return retry.OnError(retry.DefaultBackoff, retriable, func() error {
		if err := c.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete node: %w", err)
		}
		return nil
	})
}

func (c *Client) WatchNode(ctx context.Context, name string, reconcile func(node *corev1.Node) error) error {
	log := log.FromContext(ctx).WithValues("name", name)
	w, err := c.CoreV1().Nodes().Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{Name: name}))
	if err != nil {
		return fmt.Errorf("failed to watch node name=%s: %w", name, err)
	}
	defer w.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt, ok := <-w.ResultChan():
			if !ok {
				return fmt.Errorf("watch closed")
			}
			node, ok := evt.Object.(*corev1.Node)
			if !ok {
				return fmt.Errorf("expected a Node but received %#v", evt.Object)
			}

			if err := reconcile(node); err != nil {
				log.Error(err, "Reconcile Node failed")
			}
		}
	}
}

// NodeVersions returns a map of node names to their parsed Kubernetes versions.
func (c *Client) NodeVersions(ctx context.Context) (map[string]*versionutil.Version, error) {
	nodes, err := c.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	nodeVersions := make(map[string]*versionutil.Version)
	for _, node := range nodes.Items {
		v, err := versionutil.ParseGeneric(node.Status.NodeInfo.KubeletVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to parse version for node %s: %w", node.Name, err)
		}
		nodeVersions[node.Name] = v
	}

	return nodeVersions, nil
}

// CordonNode marks a node as unschedulable, preventing new pods from being scheduled on it.
func (c *Client) CordonNode(ctx context.Context, nodeName string) error {
	log := log.FromContext(ctx).WithValues("node", nodeName, "scope", "CordonNode")

	patch := []byte(`{"spec":{"unschedulable":true}}`)
	if _, err := c.CoreV1().Nodes().Patch(ctx, nodeName, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("failed to cordon node: %w", err)
	}

	log.Info("Node cordoned successfully")
	return nil
}

// DrainOpts defines options for draining a node.
type DrainOpts struct {
	// Timeout is the maximum duration to wait for the drain operation to complete.
	Timeout time.Duration
	// DeleteEmptydirData indicates whether to delete pods using emptyDir volumes.
	// Local data that will be deleted when the node is drained.
	// Equivalent to --delete-emptydir-data flag in kubectl drain.
	DeleteEmptydirData bool
	// Force indicates whether to force drain even if there are pods without controllers.
	// Equivalent to --force flag in kubectl drain.
	Force bool
	// GracePeriodSeconds period of time in seconds given to each pod to terminate gracefully.
	// If negative, the default value specified in the pod will be used.
	// Equivalent to --grace-period flag in kubectl drain.
	GracePeriodSeconds int64
	// IgnoreDaemonsets indicates whether to ignore DaemonSet-managed pods.
	// Equivalent to --ignore-daemonsets flag in kubectl drain.
	IgnoreDaemonsets bool
	// AllowDeletion indicates whether to allow deletion of pods that are blocked by PodDisruptionBudgets.
	// If true, pods that cannot be evicted due to PDB constraints will be force deleted.
	AllowDeletion bool
}

func (o DrainOpts) defaults() DrainOpts {
	return DrainOpts{
		GracePeriodSeconds: -1,
	}
}

// DrainNode drains a node by evicting all pods running on it.
func (c *Client) DrainNode(ctx context.Context, nodeName string, opts ...DrainOpts) error {
	opt := (DrainOpts{}).defaults()
	if len(opts) > 0 {
		opt = opts[0]
	}

	log := log.FromContext(ctx).WithValues("node", nodeName, "scope", "DrainNode")
	log.Info("Starting node drain")

	// List all pods on the node
	fieldSelector := fields.OneTermEqualSelector("spec.nodeName", nodeName).String()
	pods, err := c.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return fmt.Errorf("failed to list pods on node: %w", err)
	}

	var podsToEvict []corev1.Pod
	for _, pod := range pods.Items {
		// Skip pods that are already terminating
		if pod.DeletionTimestamp != nil {
			log.V(1).Info("Skipping pod that is already terminating", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}

		// Skip static pods (those managed by kubelet directly)
		if _, isStatic := pod.Annotations[corev1.MirrorPodAnnotationKey]; isStatic {
			log.V(1).Info("Skipping static pod", "pod", pod.Name, "namespace", pod.Namespace)
			continue
		}

		if opt.IgnoreDaemonsets {
			// Skip DaemonSet pods (they are managed by DaemonSets and will be recreated)
			isDaemonSet := false
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "DaemonSet" {
					isDaemonSet = true
					break
				}
			}
			if isDaemonSet {
				log.V(1).Info("Skipping DaemonSet pod", "pod", pod.Name, "namespace", pod.Namespace)
				continue
			}
		}

		if !opt.DeleteEmptydirData {
			// Do not continue if there are pods using emptyDir
			// (local data that will be deleted when the node is drained)
			for _, volume := range pod.Spec.Volumes {
				if volume.EmptyDir != nil {
					return fmt.Errorf("pod %s/%s is using emptyDir volume; cannot drain node without DeleteEmptydirData option", pod.Namespace, pod.Name)
				}
			}
		}

		if !opt.Force {
			// Stop if there are pods that do not declare a controller
			hasController := false
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Controller != nil && *ownerRef.Controller {
					hasController = true
					break
				}
			}
			if !hasController {
				return fmt.Errorf("pod %s/%s does not have a controller; cannot drain node without Force option", pod.Namespace, pod.Name)
			}
		}

		podsToEvict = append(podsToEvict, pod)
	}

	if len(podsToEvict) == 0 {
		log.Info("No pods to evict")
		return nil
	}

	log.Info("Evicting pods", "count", len(podsToEvict))

	drainCtx := ctx
	if opt.Timeout > 0 {
		var cancel context.CancelFunc
		drainCtx, cancel = context.WithTimeout(ctx, opt.Timeout)
		defer cancel()
	}

	for _, pod := range podsToEvict {
		podLog := log.WithValues("pod", pod.Name, "namespace", pod.Namespace)

		// Try to use eviction API first (respects PodDisruptionBudgets)
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &metav1.DeleteOptions{
				GracePeriodSeconds: &opt.GracePeriodSeconds,
			},
		}

		err := c.CoreV1().Pods(pod.Namespace).EvictV1(drainCtx, eviction)
		if err != nil {
			// Evictions are treated as “disruptions” that are rate-limited by a PDB.
			// When there’s no remaining budget, the API responds with 429 to signal a transient
			// condition: “try again later,” not a permanent denial.
			// 429 was chosen (instead of e.g. 403) so clients can back off and retry once budget becomes available.
			// https://kubernetes.io/docs/concepts/scheduling-eviction/api-eviction/#how-api-initiated-eviction-works
			if apierrors.IsTooManyRequests(err) {
				if opt.AllowDeletion {
					// PodDisruptionBudget is preventing eviction, force delete
					podLog.Info("Eviction blocked by PDB, force deleting pod")
					deleteOptions := metav1.DeleteOptions{
						GracePeriodSeconds: ptr.To(int64(0)),
					}
					err = c.CoreV1().Pods(pod.Namespace).Delete(drainCtx, pod.Name, deleteOptions)
					if err != nil && !apierrors.IsNotFound(err) {
						podLog.Error(err, "Failed to force delete pod")
						continue
					}
					podLog.Info("Pod force deleted successfully")
				}
			} else if !apierrors.IsNotFound(err) {
				podLog.Error(err, "Failed to evict pod")
				continue
			}
		}

		podLog.Info("Pod eviction initiated")
	}

	log.Info("Node drain completed")
	return nil
}
