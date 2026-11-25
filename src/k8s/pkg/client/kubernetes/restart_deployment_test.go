package kubernetes

import (
	"context"
	"testing"
	"time"

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
				g.Expect(err).To(Not(HaveOccurred()))
				deploy, err := client.AppsV1().Deployments("namespace").Get(context.Background(), "test", metav1.GetOptions{})
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(deploy.Spec.Template.Annotations[RestartedAtAnnotation]).NotTo(BeEmpty())
			}
		})
	}
}

func TestSetDeploymentRestartedAt(t *testing.T) {
	restartedAt := time.Date(2026, time.June, 4, 12, 30, 45, 123, time.UTC)
	for _, annotations := range []map[string]string{
		nil,
		{},
		{"other": "preserved", RestartedAtAnnotation: "old"},
	} {
		_, hasOther := annotations["other"]
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"deployment": "preserved"}},
			Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Annotations: annotations},
			}},
		}
		SetDeploymentRestartedAt(deployment, restartedAt)
		g := NewWithT(t)
		g.Expect(deployment.Spec.Template.Annotations[RestartedAtAnnotation]).To(Equal("2026-06-04T12:30:45Z"))
		g.Expect(deployment.Annotations).To(Equal(map[string]string{"deployment": "preserved"}))
		if hasOther {
			g.Expect(deployment.Spec.Template.Annotations["other"]).To(Equal("preserved"))
		}
		SetDeploymentRestartedAt(deployment, restartedAt.Add(time.Second))
		g.Expect(deployment.Spec.Template.Annotations[RestartedAtAnnotation]).To(Equal("2026-06-04T12:30:46Z"))
	}
}
