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

func TestRestartDeployment(t *testing.T) {
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
			name: "deployment",
			objects: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "namespace",
					},
				},
			},
		},
		{
			name: "deployment with other annotations",
			objects: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "namespace",
					},
					Spec: appsv1.DeploymentSpec{
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

			err := client.RestartDeployment(context.Background(), "test", "namespace")
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(BeNil())
				deploy, err := client.AppsV1().Deployments("namespace").Get(context.Background(), "test", metav1.GetOptions{})
				g.Expect(err).To(BeNil())
				g.Expect(deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"]).NotTo(BeEmpty())
			}
		})
	}
}
