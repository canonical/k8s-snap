package dnsrebalancer

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/features/coredns"
	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestReconcile_LessThanTwoNodesReady(t *testing.T) {
	g := NewWithT(t)
	r, state := newReconcileTestController(t, nil)
	ctx := context.Background()
	node := &corev1.Node{}
	g.Expect(r.client.Get(ctx, client.ObjectKey{Name: "node-2"}, node)).To(Succeed())
	g.Expect(r.client.Delete(ctx, node)).To(Succeed())
	expectReconcile(t, r, 0)
	g.Expect(state.updates).To(BeZero())
	node.ResourceVersion = ""
	g.Expect(r.client.Create(ctx, node)).To(Succeed())
	expectReconcile(t, r, 0)
	g.Expect(state.updates).To(Equal(1))
}

func TestReconcile_CoreDNSAlreadyBalanced(t *testing.T) {
	r, state := newReconcileTestController(t, func(_ *corev1.Node, pod *corev1.Pod, _ *appsv1.Deployment) {
		pod.Spec.NodeName = "node-2"
	})
	expectReconcile(t, r, 0)
	NewWithT(t).Expect(state.updates).To(BeZero())
}

func TestReconcile_RestartsDeploymentForCoLocatedPods(t *testing.T) {
	g := NewWithT(t)
	r, state := newReconcileTestController(t, nil)
	expectReconcile(t, r, 0)
	g.Expect(state.updates).To(Equal(1))
	deployment := getTestDeployment(t, r)
	g.Expect(deployment.Spec.Template.Annotations[kubernetes.RestartedAtAnnotation]).To(Equal(state.now.Format(time.RFC3339)))
}

func TestReconcile_IgnoresTerminalPods(t *testing.T) {
	for _, phase := range []corev1.PodPhase{corev1.PodFailed, corev1.PodSucceeded} {
		t.Run(string(phase), func(t *testing.T) {
			g := NewWithT(t)
			r, state := newReconcileTestController(t, nil)
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "old-coredns", Namespace: coredns.Chart.Namespace, Labels: corednsPodLabels},
				Spec:       corev1.PodSpec{NodeName: "node-2"},
				Status:     corev1.PodStatus{Phase: phase, Reason: "Evicted"},
			}
			g.Expect(r.client.Create(context.Background(), pod)).To(Succeed())
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(Equal(1))
		})
	}
}

func TestReconcile_SafetyGuards(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*corev1.Node, *corev1.Pod, *appsv1.Deployment)
	}{
		{"NodeNotReady", func(node *corev1.Node, _ *corev1.Pod, _ *appsv1.Deployment) {
			node.Status.Conditions[0].Status = corev1.ConditionFalse
		}},
		{"NodeCordoned", func(node *corev1.Node, _ *corev1.Pod, _ *appsv1.Deployment) { node.Spec.Unschedulable = true }},
		{"CNIStartupTaint", func(node *corev1.Node, _ *corev1.Pod, _ *appsv1.Deployment) {
			node.Spec.Taints = []corev1.Taint{{Key: "node.cilium.io/agent-not-ready", Effect: corev1.TaintEffectNoSchedule}}
		}},
		{"NoExecuteTaint", func(node *corev1.Node, _ *corev1.Pod, _ *appsv1.Deployment) {
			node.Spec.Taints = []corev1.Taint{{Key: "maintenance", Effect: corev1.TaintEffectNoExecute}}
		}},
		{"ControlPlaneNoScheduleTaint", func(node *corev1.Node, _ *corev1.Pod, _ *appsv1.Deployment) {
			node.Spec.Taints = []corev1.Taint{{Key: "node-role.kubernetes.io/control-plane", Effect: corev1.TaintEffectNoSchedule}}
		}},
		{"PodNotReady", func(_ *corev1.Node, pod *corev1.Pod, _ *appsv1.Deployment) {
			pod.Status.Conditions[0].Status = corev1.ConditionFalse
		}},
		{"PodPending", func(_ *corev1.Node, pod *corev1.Pod, _ *appsv1.Deployment) { pod.Status.Phase = corev1.PodPending }},
		{"UnscheduledPod", func(_ *corev1.Node, pod *corev1.Pod, _ *appsv1.Deployment) { pod.Spec.NodeName = "" }},
		{"TerminatingPod", func(_ *corev1.Node, pod *corev1.Pod, _ *appsv1.Deployment) {
			pod.DeletionTimestamp = &pod.CreationTimestamp
			pod.Finalizers = []string{"test.example/hold"}
		}},
		{"UnavailableReplica", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Status.UnavailableReplicas = 1 }},
		{"UnreadyReplica", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Status.ReadyReplicas = 1 }},
		{"UnobservedGeneration", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Generation++ }},
		{"OldReadyReplicas", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Status.UpdatedReplicas = 1 }},
		{"MinReadySecondsPending", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Status.AvailableReplicas = 1 }},
		{"ExtraReplica", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Status.Replicas = 3 }},
		{"ScaleUpPending", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { *d.Spec.Replicas = 3 }},
		{"ScaleDownPending", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { *d.Spec.Replicas = 1 }},
		{"ScaleToZeroPending", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { *d.Spec.Replicas = 0 }},
		{"DefaultReplicaCount", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) { d.Spec.Replicas = nil }},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, state := newReconcileTestController(t, test.mutate)
			expectReconcile(t, r, 0)
			NewWithT(t).Expect(state.updates).To(BeZero())
		})
	}
}

func TestReconcile_TimeGuardsWaitOnlyRemainingTime(t *testing.T) {
	for _, test := range []struct {
		name                     string
		podAge, restartAge, wait time.Duration
	}{
		{"YoungPod", 25 * time.Second, time.Minute, 5 * time.Second},
		{"RecentRestart", time.Minute, 20 * time.Second, 10 * time.Second},
		{"Both", 25 * time.Second, 20 * time.Second, 10 * time.Second},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			r, state := newReconcileTestController(t, nil)
			pod := getTestPod(t, r, "coredns-1")
			pod.CreationTimestamp = metav1.NewTime(state.now.Add(-test.podAge))
			g.Expect(r.client.Update(context.Background(), pod)).To(Succeed())
			deployment := getTestDeployment(t, r)
			kubernetes.SetDeploymentRestartedAt(deployment, state.now.Add(-test.restartAge))
			updateSettledDeployment(t, r, deployment)
			state.updates = 0
			expectReconcile(t, r, test.wait)
			state.now = state.now.Add(time.Second)
			expectReconcile(t, r, test.wait-time.Second)
			g.Expect(state.updates).To(BeZero())
			state.now = state.now.Add(test.wait - time.Second)
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(Equal(1))
		})
	}
}

func TestReconcile_MissingAndSurgePods(t *testing.T) {
	for _, count := range []int{0, 1, 3} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			g := NewWithT(t)
			r, state := newReconcileTestController(t, nil)
			ctx := context.Background()
			for i := count + 1; i <= 2; i++ {
				g.Expect(r.client.Delete(ctx, getTestPod(t, r, fmt.Sprintf("coredns-%d", i)))).To(Succeed())
			}
			if count == 3 {
				pod := getTestPod(t, r, "coredns-1")
				pod.Name, pod.ResourceVersion, pod.UID = "surge", "", "surge"
				g.Expect(r.client.Create(ctx, pod)).To(Succeed())
			}
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(BeZero())
		})
	}
}

func TestReconcile_UnsuccessfulRolloutDoesNotLoop(t *testing.T) {
	g := NewWithT(t)
	r, state := newReconcileTestController(t, nil)
	expectReconcile(t, r, 0)
	state.now = state.now.Add(time.Minute)
	// Cached pods still belong to the preceding rollout.
	expectReconcile(t, r, 0)
	g.Expect(state.updates).To(Equal(1))
	settleTestRollout(t, r, state, false)
	for range 8 {
		expectReconcile(t, r, 0)
		state.now = state.now.Add(time.Minute)
	}
	g.Expect(state.updates).To(Equal(1))
}

func TestReconcile_LaterRecoveryPermitsRebalance(t *testing.T) {
	for _, scenario := range []string{"SpreadThenColocated", "Eviction", "Upgrade", "ManualRestart", "DeploymentRecreated", "NodeReplaced", "CordonUncordon", "DNSDisabled"} {
		t.Run(scenario, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			r, state := newReconcileTestController(t, nil)
			expectReconcile(t, r, 0)
			settleTestRollout(t, r, state, scenario == "SpreadThenColocated")
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(Equal(1))
			switch scenario {
			case "SpreadThenColocated":
				pod := getTestPod(t, r, "coredns-1")
				pod.Spec.NodeName = "node-1"
				g.Expect(r.client.Update(ctx, pod)).To(Succeed())
			case "Eviction":
				pod := getTestPod(t, r, "coredns-1")
				pod.Status.Phase, pod.Status.Reason = corev1.PodFailed, "Evicted"
				g.Expect(r.client.Status().Update(ctx, pod)).To(Succeed())
				pod.Name, pod.UID, pod.ResourceVersion = "replacement", "replacement", ""
				pod.Status.Phase, pod.Status.Reason = corev1.PodRunning, ""
				g.Expect(r.client.Create(ctx, pod)).To(Succeed())
			case "Upgrade", "ManualRestart":
				deployment := getTestDeployment(t, r)
				if scenario == "Upgrade" {
					deployment.Spec.Template.Spec.Containers = []corev1.Container{{Name: "coredns", Image: "coredns:new"}}
				} else {
					kubernetes.SetDeploymentRestartedAt(deployment, state.now)
				}
				updateSettledDeployment(t, r, deployment)
				state.now = state.now.Add(time.Minute)
			case "DeploymentRecreated":
				deployment := getTestDeployment(t, r)
				g.Expect(r.client.Delete(ctx, deployment)).To(Succeed())
				deployment.UID, deployment.ResourceVersion = "replacement-deployment", ""
				g.Expect(r.client.Create(ctx, deployment)).To(Succeed())
			case "NodeReplaced":
				node := &corev1.Node{}
				g.Expect(r.client.Get(ctx, client.ObjectKey{Name: "node-2"}, node)).To(Succeed())
				g.Expect(r.client.Delete(ctx, node)).To(Succeed())
				node.UID, node.ResourceVersion = "replacement-node", ""
				g.Expect(r.client.Create(ctx, node)).To(Succeed())
			case "CordonUncordon":
				node := &corev1.Node{}
				g.Expect(r.client.Get(ctx, client.ObjectKey{Name: "node-2"}, node)).To(Succeed())
				node.Spec.Unschedulable = true
				g.Expect(r.client.Update(ctx, node)).To(Succeed())
				expectReconcile(t, r, 0)
				node.Spec.Unschedulable = false
				g.Expect(r.client.Update(ctx, node)).To(Succeed())
			case "DNSDisabled":
				getter := r.getClusterConfig
				r.getClusterConfig = func(context.Context) (types.ClusterConfig, error) { return types.ClusterConfig{}, nil }
				expectReconcile(t, r, 0)
				r.getClusterConfig = getter
			}
			before := state.updates
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(Equal(before + 1))
		})
	}
}

func TestReconcile_StatusRecoveryResumesRebalance(t *testing.T) {
	for _, scenario := range []string{"PodReady", "DeploymentReady", "ScaleUp"} {
		t.Run(scenario, func(t *testing.T) {
			g := NewWithT(t)
			r, state := newReconcileTestController(t, func(_ *corev1.Node, pod *corev1.Pod, d *appsv1.Deployment) {
				switch scenario {
				case "PodReady":
					pod.Status.Conditions[0].Status = corev1.ConditionFalse
				case "DeploymentReady":
					d.Status.AvailableReplicas = 1
				case "ScaleUp":
					*d.Spec.Replicas = 1
				}
			})
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(BeZero())
			if scenario == "PodReady" {
				pod := getTestPod(t, r, "coredns-1")
				pod.Status.Conditions[0].Status = corev1.ConditionTrue
				g.Expect(r.client.Status().Update(context.Background(), pod)).To(Succeed())
			} else {
				d := getTestDeployment(t, r)
				*d.Spec.Replicas, d.Status.AvailableReplicas = 2, 2
				updateSettledDeployment(t, r, d)
			}
			before := state.updates
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(Equal(before + 1))
		})
	}
}

func TestReconcile_AllowedRebalances(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*corev1.Node, *corev1.Pod, *appsv1.Deployment)
	}{
		{"PreferNoSchedule", func(node *corev1.Node, _ *corev1.Pod, _ *appsv1.Deployment) {
			node.Spec.Taints = []corev1.Taint{{Key: "preference", Effect: corev1.TaintEffectPreferNoSchedule}}
		}},
		{"MalformedRestartAnnotation", func(_ *corev1.Node, _ *corev1.Pod, d *appsv1.Deployment) {
			d.Spec.Template.Annotations = map[string]string{kubernetes.RestartedAtAnnotation: "invalid"}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r, state := newReconcileTestController(t, test.mutate)
			expectReconcile(t, r, 0)
			NewWithT(t).Expect(state.updates).To(Equal(1))
		})
	}
}

func TestReconcile_ClusterConfig(t *testing.T) {
	configError := errors.New("cluster config unavailable")
	for _, scenario := range []string{"NilGetter", "DNSDisabled", "ConfigError"} {
		t.Run(scenario, func(t *testing.T) {
			g := NewWithT(t)
			r, state := newReconcileTestController(t, nil)
			var want error
			switch scenario {
			case "NilGetter":
				r.getClusterConfig = nil
			case "DNSDisabled":
				r.getClusterConfig = func(context.Context) (types.ClusterConfig, error) { return types.ClusterConfig{}, nil }
			case "ConfigError":
				want = configError
				r.getClusterConfig = func(context.Context) (types.ClusterConfig, error) { return types.ClusterConfig{}, configError }
			}
			result, err := r.Reconcile(context.Background(), ctrl.Request{})
			g.Expect(errors.Is(err, want)).To(BeTrue())
			g.Expect(result).To(Equal(ctrl.Result{}))
			g.Expect(state.updates).To(BeZero())
		})
	}
}

func TestReconcile_APIErrorsDoNotConsumeAttempt(t *testing.T) {
	for _, operation := range []string{"ListNodes", "ListPods", "GetDeployment", "MissingDeployment", "UpdateConflict"} {
		t.Run(operation, func(t *testing.T) {
			g := NewWithT(t)
			r, state := newReconcileTestController(t, nil)
			original := r.client
			apiError := errors.New("API unavailable")
			if operation == "MissingDeployment" {
				apiError = apierrors.NewNotFound(appsv1.Resource("deployments"), coredns.DeploymentName)
			}
			if operation == "UpdateConflict" {
				apiError = apierrors.NewConflict(appsv1.Resource("deployments"), coredns.DeploymentName, apiError)
			}
			r.client = interceptor.NewClient(original.(client.WithWatch), interceptor.Funcs{
				List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
					_, pods := list.(*corev1.PodList)
					if (pods && operation == "ListPods") || (!pods && operation == "ListNodes") {
						return apiError
					}
					return c.List(ctx, list, opts...)
				},
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
					if operation == "GetDeployment" || operation == "MissingDeployment" {
						return apiError
					}
					return c.Get(ctx, key, obj, opts...)
				},
				Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
					if operation == "UpdateConflict" {
						return apiError
					}
					return c.Update(ctx, obj, opts...)
				},
			})
			result, err := r.Reconcile(context.Background(), ctrl.Request{})
			if operation == "MissingDeployment" {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(errors.Is(err, apiError)).To(BeTrue())
			}
			g.Expect(result).To(Equal(ctrl.Result{}))
			g.Expect(state.updates).To(BeZero())
			r.client = original
			expectReconcile(t, r, 0)
			g.Expect(state.updates).To(Equal(1))
		})
	}
}

type reconcileTestState struct {
	now     time.Time
	updates int
}

func newReconcileTestController(t *testing.T, mutate func(*corev1.Node, *corev1.Pod, *appsv1.Deployment)) (*controller, *reconcileTestState) {
	t.Helper()
	state := &reconcileTestState{now: time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)}
	enabled, replicas := true, int32(2)
	nodes := []corev1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "node-1", UID: "node-1"}, Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "node-2", UID: "node-2"}, Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}}}},
	}
	pods := make([]corev1.Pod, 2)
	for i := range pods {
		name := fmt.Sprintf("coredns-%d", i+1)
		pods[i] = corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: coredns.Chart.Namespace, UID: k8stypes.UID(name), Labels: corednsPodLabels, CreationTimestamp: metav1.NewTime(state.now.Add(-time.Minute))},
			Spec:       corev1.PodSpec{NodeName: "node-1"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning, Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
		}
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: coredns.DeploymentName, Namespace: coredns.Chart.Namespace, UID: "deployment", Generation: 1},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
		Status:     appsv1.DeploymentStatus{ObservedGeneration: 1, Replicas: 2, UpdatedReplicas: 2, ReadyReplicas: 2, AvailableReplicas: 2},
	}
	if mutate != nil {
		mutate(&nodes[1], &pods[0], deployment)
	}
	c := fake.NewClientBuilder().WithScheme(scheme.Scheme).
		WithStatusSubresource(&appsv1.Deployment{}).
		WithObjects(&nodes[0], &nodes[1], &pods[0], &pods[1], deployment).
		WithInterceptorFuncs(interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.UpdateOption) error {
				if d, ok := obj.(*appsv1.Deployment); ok {
					old := &appsv1.Deployment{}
					if err := c.Get(ctx, client.ObjectKeyFromObject(d), old); err != nil {
						return err
					}
					// The fake client does not advance generation on spec updates.
					if !reflect.DeepEqual(old.Spec, d.Spec) {
						d.Generation = old.Generation + 1
					}
					if err := c.Update(ctx, obj, opts...); err != nil {
						return err
					}
					state.updates++
					return nil
				}
				return c.Update(ctx, obj, opts...)
			},
		}).Build()
	r := &controller{client: c, now: func() time.Time { return state.now }, getClusterConfig: func(context.Context) (types.ClusterConfig, error) {
		return types.ClusterConfig{DNS: types.DNS{Enabled: &enabled}}, nil
	}}
	return r, state
}

func expectReconcile(t *testing.T, r *controller, wait time.Duration) {
	t.Helper()
	result, err := r.Reconcile(context.Background(), ctrl.Request{})
	g := NewWithT(t)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{RequeueAfter: wait}))
}

func getTestDeployment(t *testing.T, r *controller) *appsv1.Deployment {
	t.Helper()
	d := &appsv1.Deployment{}
	NewWithT(t).Expect(r.client.Get(context.Background(), client.ObjectKey{Namespace: coredns.Chart.Namespace, Name: coredns.DeploymentName}, d)).To(Succeed())
	return d
}

func getTestPod(t *testing.T, r *controller, name string) *corev1.Pod {
	t.Helper()
	pod := &corev1.Pod{}
	NewWithT(t).Expect(r.client.Get(context.Background(), client.ObjectKey{Namespace: coredns.Chart.Namespace, Name: name}, pod)).To(Succeed())
	return pod
}

func updateSettledDeployment(t *testing.T, r *controller, d *appsv1.Deployment) {
	t.Helper()
	g := NewWithT(t)
	status := d.Status
	g.Expect(r.client.Update(context.Background(), d)).To(Succeed())
	d.Status = status
	d.Status.ObservedGeneration = d.Generation
	g.Expect(r.client.Status().Update(context.Background(), d)).To(Succeed())
}

func settleTestRollout(t *testing.T, r *controller, state *reconcileTestState, spread bool) {
	t.Helper()
	g := NewWithT(t)
	ctx := context.Background()
	d := getTestDeployment(t, r)
	d.Status.ObservedGeneration = d.Generation
	g.Expect(r.client.Status().Update(ctx, d)).To(Succeed())
	for i := 1; i <= 2; i++ {
		pod := getTestPod(t, r, fmt.Sprintf("coredns-%d", i))
		g.Expect(r.client.Delete(ctx, pod)).To(Succeed())
		pod.UID += "-rollout"
		pod.ResourceVersion = ""
		pod.Annotations = d.Spec.Template.Annotations
		pod.CreationTimestamp = metav1.NewTime(state.now)
		if spread && i == 1 {
			pod.Spec.NodeName = "node-2"
		}
		g.Expect(r.client.Create(ctx, pod)).To(Succeed())
	}
	state.now = state.now.Add(time.Minute)
}
