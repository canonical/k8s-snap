package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestDeleteNode(t *testing.T) {
	g := gomega.NewWithT(t)

	t.Run("node deletion is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}
		nodeName := "test-node"
		client.CoreV1().Nodes().Create(context.TODO(), &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
		}, metav1.CreateOptions{})

		err := client.DeleteNode(context.Background(), nodeName)
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("node deletion fails", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}
		nodeName := "test-node"
		expectedErr := errors.New("some error")
		clientset.PrependReactor("delete", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, expectedErr
		})

		err := client.DeleteNode(context.Background(), nodeName)
		g.Expect(err).To(gomega.MatchError(fmt.Errorf("failed to delete node: %w", expectedErr)))
	})
}
