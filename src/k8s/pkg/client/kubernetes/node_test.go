package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	versionutil "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestDeleteNode(t *testing.T) {
	g := NewWithT(t)

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
		g.Expect(err).To(Not(HaveOccurred()))
	})

	t.Run("node does not exist is successful", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}
		nodeName := "test-node"

		err := client.DeleteNode(context.Background(), nodeName)
		g.Expect(err).To(Not(HaveOccurred()))
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
		g.Expect(err).To(MatchError(fmt.Errorf("failed to delete node: %w", expectedErr)))
	})

	t.Run("node deletion fails with internal server error", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}
		nodeName := "test-node"
		client.CoreV1().Nodes().Create(context.TODO(), &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
		}, metav1.CreateOptions{})

		expectedErr := apierrors.NewInternalError(errors.New("database is locked"))
		clientset.PrependReactor("delete", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, nil, expectedErr
		})

		err := client.DeleteNode(context.Background(), nodeName)
		g.Expect(err).To(MatchError(fmt.Errorf("failed to delete node: %w", expectedErr)))
	})

	t.Run("node deletion succeeds with internal server error", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}
		nodeName := "test-node"
		client.CoreV1().Nodes().Create(context.TODO(), &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
		}, metav1.CreateOptions{})

		expectedErr := apierrors.NewInternalError(errors.New("database is locked"))
		tries := 0
		clientset.PrependReactor("delete", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
			if tries == 3 {
				return true, nil, nil
			}
			tries++
			return true, nil, expectedErr
		})

		err := client.DeleteNode(context.Background(), nodeName)
		g.Expect(err).To(Not(HaveOccurred()))
	})
}

func TestNodeVersions(t *testing.T) {
	g := NewWithT(t)

	t.Run("returns versions for all nodes", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(
			&v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node1"},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{KubeletVersion: "v1.28.1"},
				},
			},
			&v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node2"},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{KubeletVersion: "v1.29.0"},
				},
			},
		)

		client := &Client{Interface: clientset}

		versions, err := client.NodeVersions(context.Background())
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(versions).To(HaveLen(2))

		v1, _ := versionutil.ParseGeneric("v1.28.1")
		v2, _ := versionutil.ParseGeneric("v1.29.0")

		g.Expect(versions["node1"].EqualTo(v1)).To(BeTrue())
		g.Expect(versions["node2"].EqualTo(v2)).To(BeTrue())
	})

	t.Run("returns error on invalid version", func(t *testing.T) {
		clientset := fake.NewSimpleClientset(
			&v1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: "node-bad"},
				Status: v1.NodeStatus{
					NodeInfo: v1.NodeSystemInfo{KubeletVersion: "not-a-version"},
				},
			},
		)
		client := &Client{Interface: clientset}

		_, err := client.NodeVersions(context.Background())
		g.Expect(err).To(MatchError(ContainSubstring("failed to parse version")))
	})
}
