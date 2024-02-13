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
		client := &Client{Interface: clientset}

		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
			},
		}

		_, err := client.CoreV1().Pods("test-namespace").Create(context.Background(), pod, metav1.CreateOptions{})
		g.Expect(err).To(gomega.BeNil())

		err = client.EvictPod(context.Background(), "test-namespace", "test-pod")
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("pod does not exist", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}

		err := client.EvictPod(context.Background(), "nonexistent-namespace", "nonexistent-pod")
		g.Expect(err).To(gomega.MatchError("pods \"nonexistent-pod\" not found"))
	})
}
