package kubernetes

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestRestartDaemonset(t *testing.T) {
	tests := []struct {
		name        string
		objects     []runtime.Object
		expectError bool
	}{
		{
			name:        "missing",
			expectError: true,
		},
		{
			name: "daemonset",
			objects: []runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "namespace",
					},
				},
			},
		},
		{
			name: "daemonset with other annotations",
			objects: []runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "namespace",
					},
					Spec: appsv1.DaemonSetSpec{
						Template: corev1.PodTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Annotations: map[string]string{
									"test": "val",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			clientset := fake.NewSimpleClientset(tc.objects...)
			client := &Client{Interface: clientset}

			err := client.RestartDaemonset(context.Background(), "test", "namespace")
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
				ds, err := client.AppsV1().DaemonSets("namespace").Get(context.Background(), "test", metav1.GetOptions{})
				g.Expect(err).To(BeNil())
				g.Expect(ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]).NotTo(BeEmpty())
			}
		})
	}
}
