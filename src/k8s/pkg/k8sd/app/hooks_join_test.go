package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestOnPostJoin_DuplicateNodeName(t *testing.T) {
	g := NewWithT(t)

	t.Run("fails when node with same name already exists", func(t *testing.T) {
		// Test 1: Create fake k8s clinetset with an existing node that has the same
		// name as the joining node.
		nodeName := "test-node"
		existingNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{Type: v1.NodeReady, Status: v1.ConditionTrue},
				},
			},
		}

		clientset := fake.NewSimpleClientset(existingNode)
		k8sClient := &kubernetes.Client{Interface: clientset}

		// check that GetNode() finds the existing node.
		node, err := k8sClient.GetNode(context.Background(), nodeName)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(node.Name).To(Equal(nodeName))

		if _, err := k8sClient.GetNode(context.Background(), nodeName); err == nil {
			err = fmt.Errorf("A node with the same name %q is already part of the cluster", nodeName)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("A node with the same name %q is already part of the cluster", nodeName)))
		}
	})

	t.Run("succeeds when node with same name does not exist", func(t *testing.T) {
		// Test 2: Create fake k8s clientset with no existing nodes. Joining node
		// should succeed.
		nodeName := "new-node"
		clientset := fake.NewSimpleClientset()
		k8sClient := &kubernetes.Client{Interface: clientset}

		// Check that GetNode() returns an error indicating that no node exists.
		_, err := k8sClient.GetNode(context.Background(), nodeName)
		g.Expect(err).To(HaveOccurred())

		var joinErr error
		if _, err := k8sClient.GetNode(context.Background(), nodeName); err == nil {
			joinErr = fmt.Errorf("A node with the same name %q is already part of the cluster", nodeName)
		}
		g.Expect(joinErr).To(Not(HaveOccurred()))
	})

	t.Run("fails when worker node with same name already exists", func(t *testing.T) {
		// Test 3: Create fake k8s clientset with an existing worker node that has the same
		// name as the joining worker node.
		nodeName := "worker-node"
		existingWorkerNode := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeName,
				Labels: map[string]string{
					"node-role.kubernetes.io/worker": "",
				},
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{Type: v1.NodeReady, Status: v1.ConditionTrue},
				},
			},
		}

		clientset := fake.NewSimpleClientset(existingWorkerNode)
		k8sClient := &kubernetes.Client{Interface: clientset}

		// Check that GetNode() finds the existing worker node.
		node, err := k8sClient.GetNode(context.Background(), nodeName)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(node.Name).To(Equal(nodeName))
		g.Expect(node.Labels).To(HaveKey("node-role.kubernetes.io/worker"))

		if _, err := k8sClient.GetNode(context.Background(), nodeName); err == nil {
			err = fmt.Errorf("A node with the same name %q is already part of the cluster", nodeName)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("A node with the same name %q is already part of the cluster", nodeName)))
		}
	})
}
