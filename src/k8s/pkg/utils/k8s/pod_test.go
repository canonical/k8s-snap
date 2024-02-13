package k8s

import (
	"context"
	"testing"

	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestEvictPod(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("pod eviction is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		k8sClient := &k8sClient{Interface: clientset}

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
			},
		}

		_, err := k8sClient.CoreV1().Pods("test-namespace").Create(context.Background(), pod, metav1.CreateOptions{})
		g.Expect(err).To(gomega.BeNil())

		err = EvictPod(context.Background(), k8sClient, "test-namespace", "test-pod")
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("pod does not exist", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		k8sClient := &k8sClient{Interface: clientset}

		err := EvictPod(context.Background(), k8sClient, "nonexistent-namespace", "nonexistent-pod")
		g.Expect(err).To(gomega.MatchError("pods \"nonexistent-pod\" not found"))
	})
}
