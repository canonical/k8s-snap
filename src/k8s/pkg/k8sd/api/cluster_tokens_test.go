package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestCheckNodeNameAvailable_Duplicate_Node(t *testing.T) {
	// Tests that checkNodeNameAvailable returns error when a node with the given name already exists.
	g := NewWithT(t)
	sharedName := "duplicate-node"
	existingNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: sharedName,
		},
	}
	fakeClientset := fake.NewSimpleClientset(existingNode)
	k8sClient := &kubernetes.Client{Interface: fakeClientset}

	err := checkNodeNameAvailable(context.Background(), k8sClient, sharedName)
	g.Expect(err).To(HaveOccurred())

	g.Expect(errors.Is(err, errNodeNameAlreadyExists)).To(BeTrue())
}

func TestCheckNodeNameAvailable_Unique_Name(t *testing.T) {
	// Tests that checkNodeNameAvailable returns nil when the node name is unique.
	g := NewWithT(t)

	existingNode := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unique-node-1",
		},
	}
	fakeClientset := fake.NewSimpleClientset(existingNode)
	k8sClient := &kubernetes.Client{Interface: fakeClientset}

	// Check for a different node name
	err := checkNodeNameAvailable(context.Background(), k8sClient, "unique-node-2")
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCheckNodeNameAvailable_API_Errors(t *testing.T) {
	// Tests that checkNodeNameAvailable returns error when the K8s API throws a non-NotFound error.
	g := NewWithT(t)

	// Create fake K8s client and inject an error for GET nodes RPC.
	fakeClientset := fake.NewSimpleClientset()
	mockError := fmt.Errorf("mock API error")
	fakeClientset.PrependReactor("get", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
		_, ok := action.(k8stesting.GetAction)
		if ok {
			return true, nil, mockError
		}
		return false, nil, nil
	})

	k8sClient := &kubernetes.Client{Interface: fakeClientset}

	failedNodeName := "any-node"
	err := checkNodeNameAvailable(context.Background(), k8sClient, failedNodeName)
	g.Expect(err).To(HaveOccurred())

	g.Expect(errors.Is(err, errFailedToCheckNodeName)).To(BeTrue())
	g.Expect(errors.Is(err, mockError)).To(BeTrue())
}
