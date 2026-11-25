package dnsrebalancer

import (
	"context"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestNodeEventPredicateReadiness(t *testing.T) {
	states := []struct {
		name   string
		status corev1.ConditionStatus
		ready  bool
	}{
		{"True", corev1.ConditionTrue, true},
		{"False", corev1.ConditionFalse, false},
		{"Unknown", corev1.ConditionUnknown, false},
		{"Missing", "", false},
	}
	newNode := func(status corev1.ConditionStatus) *corev1.Node {
		node := &corev1.Node{}
		if status != "" {
			node.Status.Conditions = []corev1.NodeCondition{{Type: corev1.NodeReady, Status: status}}
		}
		return node
	}
	filter := nodeEventPredicate()
	for _, state := range states {
		t.Run("Create/"+state.name, func(t *testing.T) {
			NewWithT(t).Expect(filter.Create(event.CreateEvent{Object: newNode(state.status)})).To(Equal(state.ready))
		})
		for _, next := range states {
			t.Run("Update/"+state.name+"To"+next.name, func(t *testing.T) {
				update := event.UpdateEvent{ObjectOld: newNode(state.status), ObjectNew: newNode(next.status)}
				NewWithT(t).Expect(filter.Update(update)).To(Equal(state.ready != next.ready))
			})
		}
	}
}

func TestNodeEventPredicateUpdates(t *testing.T) {
	baseline := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Spec: corev1.NodeSpec{Taints: []corev1.Taint{
			{Key: "preference", Effect: corev1.TaintEffectPreferNoSchedule},
			{Key: "another-preference", Effect: corev1.TaintEffectPreferNoSchedule},
		}},
		Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
		}},
	}
	cases := []struct {
		name   string
		mutate func(*corev1.Node)
		want   bool
	}{
		{"Unchanged", func(node *corev1.Node) {}, false},
		{"Cordon", func(node *corev1.Node) { node.Spec.Unschedulable = true }, true},
		{"RemoveAllTaints", func(node *corev1.Node) { node.Spec.Taints = nil }, false},
		{"RemovePreferredTaint", func(node *corev1.Node) { node.Spec.Taints = node.Spec.Taints[1:] }, false},
		{"AddControlPlaneTaint", func(node *corev1.Node) {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{Key: "node-role.kubernetes.io/control-plane", Effect: corev1.TaintEffectNoSchedule})
		}, true},
		{"AddCNIStartupTaint", func(node *corev1.Node) {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{Key: "node.cilium.io/agent-not-ready", Effect: corev1.TaintEffectNoSchedule})
		}, true},
		{"AddNoExecuteTaint", func(node *corev1.Node) {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{Key: "maintenance", Effect: corev1.TaintEffectNoExecute})
		}, true},
		{"TaintKey", func(node *corev1.Node) { node.Spec.Taints[1].Key = "maintenance" }, false},
		{"TaintValue", func(node *corev1.Node) { node.Spec.Taints[1].Value = "pending" }, false},
		{"TaintEffect", func(node *corev1.Node) { node.Spec.Taints[1].Effect = corev1.TaintEffectNoExecute }, true},
		{"TaintOrder", func(node *corev1.Node) {
			node.Spec.Taints[0], node.Spec.Taints[1] = node.Spec.Taints[1], node.Spec.Taints[0]
		}, false},
		{"TaintTimeAdded", func(node *corev1.Node) {
			timestamp := metav1.NewTime(time.Unix(100, 0))
			node.Spec.Taints[0].TimeAdded = &timestamp
		}, false},
		{"Labels", func(node *corev1.Node) { node.Labels = map[string]string{"example.com/label": "value"} }, false},
		{"Annotations", func(node *corev1.Node) { node.Annotations = map[string]string{"example.com/note": "value"} }, false},
		{"ResourceVersion", func(node *corev1.Node) { node.ResourceVersion = "2" }, false},
		{"Heartbeat", func(node *corev1.Node) {
			node.Status.Conditions[0].LastHeartbeatTime = metav1.NewTime(time.Unix(100, 0))
		}, false},
		{"ReadyConditionDetails", func(node *corev1.Node) {
			node.Status.Conditions[0].Reason = "KubeletReady"
			node.Status.Conditions[0].Message = "ready"
			node.Status.Conditions[0].LastTransitionTime = metav1.NewTime(time.Unix(100, 0))
		}, false},
		{"OtherCondition", func(node *corev1.Node) {
			node.Status.Conditions = append(node.Status.Conditions, corev1.NodeCondition{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionTrue})
		}, false},
		{"Images", func(node *corev1.Node) {
			node.Status.Images = []corev1.ContainerImage{{Names: []string{"example.com/image:latest"}}}
		}, false},
		{"CombinedEligibilityChanges", func(node *corev1.Node) {
			node.Spec.Unschedulable = true
			node.Spec.Taints = nil
			node.Status.Conditions[0].Status = corev1.ConditionFalse
		}, true},
	}
	filter := nodeEventPredicate()
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			original := baseline.DeepCopy()
			changed := original.DeepCopy()
			test.mutate(changed)
			for _, direction := range []struct {
				name string
				old  *corev1.Node
				new  *corev1.Node
			}{
				{"Forward", original, changed},
				{"Reverse", changed, original},
			} {
				t.Run(direction.name, func(t *testing.T) {
					NewWithT(t).Expect(filter.Update(event.UpdateEvent{ObjectOld: direction.old, ObjectNew: direction.new})).To(Equal(test.want))
					NewWithT(t).Expect(filter.Create(event.CreateEvent{Object: direction.new})).To(Equal(nodeSchedulableForCoreDNS(direction.new)))
				})
			}
		})
	}
	t.Run("NilAndEmptyTaints", func(t *testing.T) {
		original, changed := baseline.DeepCopy(), baseline.DeepCopy()
		original.Spec.Taints = nil
		changed.Spec.Taints = []corev1.Taint{}
		NewWithT(t).Expect(filter.Update(event.UpdateEvent{ObjectOld: original, ObjectNew: changed})).To(BeFalse())
	})
}

func TestNodeEventPredicateInvalidObjects(t *testing.T) {
	filter := nodeEventPredicate()
	for _, test := range []struct {
		name   string
		object client.Object
	}{
		{"Nil", nil},
		{"NonNode", &corev1.Pod{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(filter.Create(event.CreateEvent{Object: test.object})).To(BeFalse())
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: test.object, ObjectNew: &corev1.Node{}})).To(BeFalse())
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: &corev1.Node{}, ObjectNew: test.object})).To(BeFalse())
			g.Expect(filter.Delete(event.DeleteEvent{Object: test.object})).To(BeFalse())
		})
	}
}

func TestNodeEventPredicateLifecycle(t *testing.T) {
	filter := nodeEventPredicate()
	for _, ready := range []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionUnknown} {
		t.Run(string(ready), func(t *testing.T) {
			g := NewWithT(t)
			node := &corev1.Node{Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: ready}}}}
			g.Expect(filter.Delete(event.DeleteEvent{Object: node})).To(BeTrue())
			g.Expect(filter.Delete(event.DeleteEvent{Object: node, DeleteStateUnknown: true})).To(BeTrue())
			g.Expect(filter.Generic(event.GenericEvent{Object: node})).To(BeFalse())
		})
	}
}

func newCoreDNSPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: coredns.Chart.Namespace,
			Name:      "coredns-1",
			Labels:    map[string]string{"k8s-app": "coredns", "app.kubernetes.io/instance": "ck-dns"},
		},
		Spec:   corev1.PodSpec{NodeName: "node-1"},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
	}
}

func TestCoreDNSPodEventPredicateSelection(t *testing.T) {
	otherPod := func() *corev1.Pod {
		pod := newCoreDNSPod()
		pod.Labels = map[string]string{"k8s-app": "kube-proxy"}
		return pod
	}
	wrongNamespace := func() *corev1.Pod {
		pod := newCoreDNSPod()
		pod.Namespace = "default"
		return pod
	}
	filter := coreDNSPodEventPredicate()
	for _, test := range []struct {
		name   string
		object client.Object
		want   bool
	}{
		{"CoreDNSPod", newCoreDNSPod(), true},
		{"OtherLabels", otherPod(), false},
		{"WrongNamespace", wrongNamespace(), false},
		{"Nil", nil, false},
		{"NonPod", &corev1.Node{}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(filter.Create(event.CreateEvent{Object: test.object})).To(Equal(test.want))
			g.Expect(filter.Delete(event.DeleteEvent{Object: test.object})).To(Equal(test.want))
			g.Expect(filter.Generic(event.GenericEvent{Object: test.object})).To(BeFalse())
		})
	}
}

func TestCoreDNSPodEventPredicateUpdates(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*corev1.Pod)
		want   bool
	}{
		{"Unchanged", func(*corev1.Pod) {}, false},
		{"Scheduled", func(pod *corev1.Pod) { pod.Spec.NodeName = "node-2" }, true},
		{"Ready", func(pod *corev1.Pod) { pod.Status.Conditions[0].Status = corev1.ConditionFalse }, true},
		{"Failed", func(pod *corev1.Pod) { pod.Status.Phase = corev1.PodFailed }, true},
		{"Succeeded", func(pod *corev1.Pod) { pod.Status.Phase = corev1.PodSucceeded }, true},
		{"Pending", func(pod *corev1.Pod) { pod.Status.Phase = corev1.PodPending }, true},
		{"SelectorRemoved", func(pod *corev1.Pod) { pod.Labels = nil }, true},
		{"Terminating", func(pod *corev1.Pod) {
			now := metav1.Now()
			pod.DeletionTimestamp = &now
		}, true},
		{"ContainerStatus", func(pod *corev1.Pod) {
			pod.Status.ContainerStatuses = []corev1.ContainerStatus{{Name: "coredns", Ready: true}}
		}, false},
		{"ResourceVersion", func(pod *corev1.Pod) { pod.ResourceVersion = "2" }, false},
		{"Annotations", func(pod *corev1.Pod) { pod.Annotations = map[string]string{"example.com/note": "value"} }, false},
	}
	filter := coreDNSPodEventPredicate()
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			original := newCoreDNSPod()
			changed := original.DeepCopy()
			test.mutate(changed)
			g := NewWithT(t)
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: original, ObjectNew: changed})).To(Equal(test.want))
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: changed, ObjectNew: original})).To(Equal(test.want))
		})
	}
	t.Run("NotCoreDNSPod", func(t *testing.T) {
		g := NewWithT(t)
		pod := newCoreDNSPod()
		pod.Labels = nil
		changed := pod.DeepCopy()
		changed.Spec.NodeName = "node-2"
		g.Expect(filter.Update(event.UpdateEvent{ObjectOld: pod, ObjectNew: changed})).To(BeFalse())
	})
	t.Run("InvalidObjects", func(t *testing.T) {
		g := NewWithT(t)
		pod := newCoreDNSPod()
		g.Expect(filter.Update(event.UpdateEvent{ObjectOld: &corev1.Node{}, ObjectNew: pod})).To(BeFalse())
		g.Expect(filter.Update(event.UpdateEvent{ObjectOld: pod, ObjectNew: &corev1.Node{}})).To(BeFalse())
	})
}

func TestCoreDNSDeploymentEventPredicate(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Namespace: coredns.Chart.Namespace,
		Name:      coredns.DeploymentName,
	}}
	filter := coreDNSDeploymentEventPredicate()
	for _, test := range []struct {
		name   string
		object client.Object
		want   bool
	}{
		{"CoreDNS", deployment, true},
		{"OtherDeployment", &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: coredns.Chart.Namespace, Name: "other"}}, false},
		{"WrongNamespace", &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: coredns.DeploymentName}}, false},
		{"NonDeployment", newCoreDNSPod(), false},
		{"Nil", nil, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(filter.Create(event.CreateEvent{Object: test.object})).To(Equal(test.want))
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: test.object, ObjectNew: test.object})).To(Equal(test.want))
			g.Expect(filter.Delete(event.DeleteEvent{Object: test.object})).To(Equal(test.want))
			g.Expect(filter.Generic(event.GenericEvent{Object: test.object})).To(BeFalse())
		})
	}
	t.Run("StatusOnly", func(t *testing.T) {
		changed := deployment.DeepCopy()
		changed.Status.ReadyReplicas = 2
		NewWithT(t).Expect(filter.Update(event.UpdateEvent{ObjectOld: deployment, ObjectNew: changed})).To(BeTrue())
	})
	t.Run("Spec", func(t *testing.T) {
		changed := deployment.DeepCopy()
		changed.Generation++
		changed.Spec.Template.Annotations = map[string]string{"rollout": "new"}
		NewWithT(t).Expect(filter.Update(event.UpdateEvent{ObjectOld: deployment, ObjectNew: changed})).To(BeTrue())
	})
}

func TestCoreDNSRequestHandlerUsesOneWorkKey(t *testing.T) {
	g := NewWithT(t)
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[ctrl.Request]())
	defer queue.ShutDown()
	ctx := context.Background()
	handler := coreDNSRequestHandler()
	objects := []client.Object{
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-1"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-2"}},
		newCoreDNSPod(),
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: coredns.Chart.Namespace, Name: coredns.DeploymentName}},
	}
	for _, object := range objects {
		handler.Create(ctx, event.CreateEvent{Object: object}, queue)
		handler.Update(ctx, event.UpdateEvent{ObjectOld: object, ObjectNew: object}, queue)
		handler.Delete(ctx, event.DeleteEvent{Object: object}, queue)
	}
	g.Expect(queue.Len()).To(Equal(1))
	request, shutdown := queue.Get()
	g.Expect(shutdown).To(BeFalse())
	g.Expect(request).To(Equal(ctrl.Request{NamespacedName: k8stypes.NamespacedName{
		Namespace: "kube-system",
		Name:      "coredns",
	}}))
	// Events arriving during a reconcile coalesce into one subsequent reconcile.
	for _, object := range objects {
		handler.Create(ctx, event.CreateEvent{Object: object}, queue)
	}
	g.Expect(queue.Len()).To(BeZero())
	queue.Done(request)
	g.Expect(queue.Len()).To(Equal(1))
	next, shutdown := queue.Get()
	g.Expect(shutdown).To(BeFalse())
	g.Expect(next).To(Equal(request))
	queue.Done(next)
	g.Expect(queue.Len()).To(BeZero())
}

func TestNodeEventPredicateIgnoresChangesWhileIneligible(t *testing.T) {
	filter := nodeEventPredicate()
	for _, test := range []struct {
		name   string
		mutate func(*corev1.Node)
	}{
		{"ReadyButCordoned", func(node *corev1.Node) {
			node.Status.Conditions = []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}
		}},
		{"DifferentBlockingTaint", func(node *corev1.Node) {
			node.Spec.Taints = []corev1.Taint{{Key: "maintenance", Effect: corev1.TaintEffectNoSchedule}}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			original := &corev1.Node{Spec: corev1.NodeSpec{Unschedulable: true}}
			changed := original.DeepCopy()
			test.mutate(changed)
			g := NewWithT(t)
			g.Expect(filter.Create(event.CreateEvent{Object: changed})).To(BeFalse())
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: original, ObjectNew: changed})).To(BeFalse())
			g.Expect(filter.Update(event.UpdateEvent{ObjectOld: changed, ObjectNew: original})).To(BeFalse())
		})
	}
}
