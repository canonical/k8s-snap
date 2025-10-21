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

func TestCordonNode(t *testing.T) {
	g := NewWithT(t)

	t.Run("successfully cordons node", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
			Spec:       v1.NodeSpec{Unschedulable: false},
		}
		clientset := fake.NewSimpleClientset(node)
		client := &Client{Interface: clientset}

		err := client.CordonNode(context.Background(), "test-node")
		g.Expect(err).To(Not(HaveOccurred()))

		updatedNode, err := client.CoreV1().Nodes().Get(context.Background(), "test-node", metav1.GetOptions{})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(updatedNode.Spec.Unschedulable).To(BeTrue())
	})

	t.Run("fails when node does not exist", func(t *testing.T) {
		clientset := fake.NewSimpleClientset()
		client := &Client{Interface: clientset}

		err := client.CordonNode(context.Background(), "nonexistent-node")
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("failed to cordon node"))
	})
}

func TestDrainNode(t *testing.T) {
	g := NewWithT(t)

	t.Run("successfully drains node with no pods", func(t *testing.T) {
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
		}
		clientset := fake.NewSimpleClientset(node)
		client := &Client{Interface: clientset}

		err := client.DrainNode(context.Background(), "test-node")
		g.Expect(err).To(Not(HaveOccurred()))

		// Verify no pods exist
		pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(pods.Items).To(BeEmpty())
	})

	t.Run("skips static and daemonset pods", func(t *testing.T) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		staticPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "static-pod",
				Namespace:   "default",
				Annotations: map[string]string{v1.MirrorPodAnnotationKey: "mirror"},
			},
			Spec: v1.PodSpec{NodeName: "test-node"},
		}
		daemonsetPod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daemonset-pod",
				Namespace: "kube-system",
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "DaemonSet", Name: "test-ds"},
				},
			},
			Spec: v1.PodSpec{NodeName: "test-node"},
		}

		clientset := fake.NewSimpleClientset(node, staticPod, daemonsetPod)
		client := &Client{Interface: clientset}

		// Track eviction calls
		evictionCalled := false
		clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			if action.GetSubresource() == "eviction" {
				evictionCalled = true
				return true, nil, nil
			}
			return false, nil, nil
		})

		err := client.DrainNode(context.Background(), "test-node", DrainOpts{IgnoreDaemonsets: true})
		g.Expect(err).To(Not(HaveOccurred()))

		// Verify eviction was NOT called (static and daemonset pods are skipped)
		g.Expect(evictionCalled).To(BeFalse())

		// Verify static and daemonset pods still exist
		pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(pods.Items).To(HaveLen(2))

		podNames := make(map[string]struct{})
		for _, pod := range pods.Items {
			podNames[pod.Name] = struct{}{}
		}
		g.Expect(podNames).To(HaveKey("static-pod"))
		g.Expect(podNames).To(HaveKey("daemonset-pod"))
	})

	t.Run("fails when pod uses emptyDir without DeleteEmptydirData", func(t *testing.T) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod-with-emptydir",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "ReplicaSet", Controller: boolPtr(true)},
				},
			},
			Spec: v1.PodSpec{
				NodeName: "test-node",
				Volumes:  []v1.Volume{{Name: "data", VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{}}}},
			},
		}

		clientset := fake.NewSimpleClientset(node, pod)
		client := &Client{Interface: clientset}

		// Track eviction calls
		evictionCalled := false
		clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			if action.GetSubresource() == "eviction" {
				evictionCalled = true
				return true, nil, nil
			}
			return false, nil, nil
		})

		err := client.DrainNode(context.Background(), "test-node")
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("using emptyDir volume"))

		// Verify eviction was NOT called (drain failed before attempting eviction)
		g.Expect(evictionCalled).To(BeFalse())

		// Verify pod still exists (not drained)
		remainingPod, err := client.CoreV1().Pods("default").Get(context.Background(), "pod-with-emptydir", metav1.GetOptions{})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(remainingPod.Name).To(Equal("pod-with-emptydir"))
	})

	t.Run("fails when pod has no controller without Force", func(t *testing.T) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "standalone-pod",
				Namespace: "default",
			},
			Spec: v1.PodSpec{NodeName: "test-node"},
		}

		clientset := fake.NewSimpleClientset(node, pod)
		client := &Client{Interface: clientset}

		// Track eviction calls
		evictionCalled := false
		clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			if action.GetSubresource() == "eviction" {
				evictionCalled = true
				return true, nil, nil
			}
			return false, nil, nil
		})

		err := client.DrainNode(context.Background(), "test-node")
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("does not have a controller"))

		// Verify eviction was NOT called (drain failed before attempting eviction)
		g.Expect(evictionCalled).To(BeFalse())

		// Verify pod still exists (not drained)
		remainingPod, err := client.CoreV1().Pods("default").Get(context.Background(), "standalone-pod", metav1.GetOptions{})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(remainingPod.Name).To(Equal("standalone-pod"))
	})

	t.Run("evicts pods with controller", func(t *testing.T) {
		node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "test-node"}}
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "managed-pod",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					{Kind: "ReplicaSet", Controller: boolPtr(true)},
				},
			},
			Spec: v1.PodSpec{NodeName: "test-node"},
		}

		clientset := fake.NewSimpleClientset(node, pod)
		client := &Client{Interface: clientset}

		// Track eviction calls
		evictionCalled := false
		clientset.PrependReactor("create", "pods", func(action k8stesting.Action) (bool, runtime.Object, error) {
			if action.GetSubresource() == "eviction" {
				evictionCalled = true
				// Simulate successful eviction by returning no error
				return true, nil, nil
			}
			return false, nil, nil
		})

		err := client.DrainNode(context.Background(), "test-node")
		g.Expect(err).To(Not(HaveOccurred()))

		// Verify eviction was called
		g.Expect(evictionCalled).To(BeTrue())
	})
}

func boolPtr(b bool) *bool {
	return &b
}
