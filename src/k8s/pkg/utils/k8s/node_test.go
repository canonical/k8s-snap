package k8s

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

func TestDrainNode(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("when draining a node is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"spec.nodeName": "test-node",
				},
			},
		})
		client := &k8sClient{Interface: clientset}

		err := DrainNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("when getting pods for node fails", func(t *testing.T) {
		expectedErr := errors.New("some error")
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("list", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, expectedErr
		})
		client := &k8sClient{Interface: clientset}

		err := DrainNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error())))
	})
}

func TestCordonNode(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("when cordon node is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		})
		client := &k8sClient{Interface: clientset}

		err := CordonNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("when getting node fails", func(t *testing.T) {
		expectedErr := errors.New("some error")
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, expectedErr
		})
		client := &k8sClient{Interface: clientset}

		err := CordonNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error())))
	})

	t.Run("when updating node fails", func(t *testing.T) {
		expectedErr := errors.New("some error")
		clientset := fake.NewSimpleClientset(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		})
		clientset.PrependReactor("update", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, expectedErr
		})
		client := &k8sClient{Interface: clientset}

		err := CordonNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error())))
	})
}

func TestUncordonNode(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("when uncordon node is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		})
		client := &k8sClient{Interface: clientset}

		err := UncordonNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("when getting node fails", func(t *testing.T) {
		expectedErr := errors.New("some error")
		clientset := fake.NewSimpleClientset()
		clientset.PrependReactor("get", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, expectedErr
		})
		client := &k8sClient{Interface: clientset}

		err := UncordonNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error())))
	})

	t.Run("when updating node fails", func(t *testing.T) {
		expectedErr := errors.New("some error")
		clientset := fake.NewSimpleClientset(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		})
		clientset.PrependReactor("update", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, expectedErr
		})
		client := &k8sClient{Interface: clientset}

		err := UncordonNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error())))
	})
}

func TestDeleteNode(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("node deletion is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		k8sClient := &k8sClient{Interface: clientset}
		nodeName := "test-node"
		k8sClient.CoreV1().Nodes().Create(context.TODO(), &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
		}, metav1.CreateOptions{})

		err := DeleteNode(context.Background(), k8sClient, nodeName)
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("node deletion fails", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		k8sClient := &k8sClient{Interface: clientset}
		nodeName := "test-node"
		expectedErr := errors.New("some error")
		clientset.PrependReactor("delete", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, expectedErr
		})

		err := DeleteNode(context.Background(), k8sClient, nodeName)
		g.Expect(err).To(gomega.MatchError(fmt.Errorf("failed to delete node: %w", expectedErr)))
	})
}

func TestGracefullyDeleteNode(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	t.Run("when gracefully deleting a node is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		}, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"spec.nodeName": "test-node",
				},
			},
		})
		client := &k8sClient{Interface: clientset}

		err := GracefullyDeleteNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.BeNil())
	})

	t.Run("when deleting node fails", func(t *testing.T) {
		expectedErr := errors.New("some error")
		clientset := fake.NewSimpleClientset(&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		}, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					"spec.nodeName": "test-node",
				},
			},
		})
		clientset.PrependReactor("delete", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, expectedErr
		})
		client := &k8sClient{Interface: clientset}

		err := GracefullyDeleteNode(context.Background(), client, "test-node")
		g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedErr.Error())))
	})
}
