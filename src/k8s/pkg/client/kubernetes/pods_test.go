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
	g := NewGomegaWithT(t)

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
				g.Expect(err).Should(BeNil())
			} else {
				g.Expect(err).Should(MatchError(tc.expectedError))
			}
		})
	}
}
