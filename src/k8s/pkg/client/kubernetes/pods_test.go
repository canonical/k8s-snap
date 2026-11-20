package kubernetes

import (
	"context"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestCheckForReadyPods(t *testing.T) {
	testCases := []struct {
		name          string
		namespace     string
		listOptions   metav1.ListOptions
		podList       *corev1.PodList
		listError     error
		expectedError string
	}{
		{
			name:          "No pods",
			namespace:     "test-namespace",
			podList:       &corev1.PodList{},
			expectedError: "no pods in test-namespace namespace on the cluster",
		},
		{
			name:      "All pods ready",
			namespace: "test-namespace",
			podList: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
							Conditions: []corev1.PodCondition{
								{Type: corev1.PodReady, Status: corev1.ConditionTrue},
							},
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name:      "Some pods not ready",
			namespace: "test-namespace",
			podList: &corev1.PodList{
				Items: []corev1.Pod{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod1"},
						Status: corev1.PodStatus{
							Phase: corev1.PodRunning,
							Conditions: []corev1.PodCondition{
								{Type: corev1.PodReady, Status: corev1.ConditionTrue},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "pod2"},
						Status: corev1.PodStatus{
							Phase: corev1.PodPending,
						},
					},
				},
			},
			expectedError: "pods [pod2] not ready",
		},
		{
			name:          "Error listing pods",
			namespace:     "test-namespace",
			listError:     fmt.Errorf("list error"),
			expectedError: "failed to list pods: list error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			clientset := fake.NewSimpleClientset()
			client := &Client{
				Interface: clientset,
			}

			// Setup fake client responses
			clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
				if tc.listError != nil {
					return true, nil, tc.listError
				}
				return true, tc.podList, nil
			})

			err := client.CheckForReadyPods(context.Background(), tc.namespace, tc.listOptions)

			if tc.expectedError == "" {
				g.Expect(err).ShouldNot(HaveOccurred())
			} else {
				g.Expect(err).Should(MatchError(tc.expectedError))
			}
		})
	}
}

func TestIsPodReady(t *testing.T) {
	if IsPodReady(nil) {
		t.Fatal("nil pod must not be ready")
	}
	for _, phase := range []corev1.PodPhase{"", corev1.PodPending, corev1.PodRunning, corev1.PodSucceeded, corev1.PodFailed, corev1.PodUnknown} {
		for _, status := range []corev1.ConditionStatus{corev1.ConditionTrue, corev1.ConditionFalse, corev1.ConditionUnknown} {
			t.Run(string(phase)+"/"+string(status), func(t *testing.T) {
				pod := &corev1.Pod{Status: corev1.PodStatus{
					Phase: phase,
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodScheduled, Status: corev1.ConditionTrue},
						{Type: corev1.PodReady, Status: status},
					},
				}}
				want := phase == corev1.PodRunning && status == corev1.ConditionTrue
				if got := IsPodReady(pod); got != want {
					t.Fatalf("IsPodReady() = %v, want %v", got, want)
				}
			})
		}
	}
	pod := &corev1.Pod{Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	if IsPodReady(pod) {
		t.Fatal("running pod without conditions must not be ready")
	}
	pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodScheduled, Status: corev1.ConditionTrue}}
	if IsPodReady(pod) {
		t.Fatal("running pod without a Ready condition must not be ready")
	}
}
