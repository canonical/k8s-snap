package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	snapmock "github.com/canonical/k8s/pkg/snap/mock"
	"github.com/canonical/lxd/shared/revert"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestRegisterK8sDqliteReverter tests that the k8s-dqlite reverter properly cleans up state.
func TestRegisterK8sDqliteReverter(t *testing.T) {
	g := NewWithT(t)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	dqliteStateDir := filepath.Join(tmpDir, "dqlite")
	g.Expect(os.MkdirAll(dqliteStateDir, 0o755)).To(Succeed())

	// Create a test file in the state directory
	testFile := filepath.Join(dqliteStateDir, "test.db")
	g.Expect(os.WriteFile(testFile, []byte("test data"), 0o644)).To(Succeed())

	// Verify file exists before cleanup
	g.Expect(testFile).To(BeAnExistingFile())

	// Create mock snap and reverter
	mockSnap := &snapmock.Snap{
		Mock: snapmock.Mock{
			K8sDqliteStateDir: dqliteStateDir,
		},
	}
	reverter := revert.New()

	// Register the reverter
	registerK8sDqliteReverter(mockSnap, reverter)

	// Trigger the reverter (simulating join failure)
	reverter.Fail()

	// Verify the state directory was cleaned up
	g.Expect(dqliteStateDir).NotTo(BeAnExistingFile())
}

// TestRegisterK8sDqliteReverter_Success tests that cleanup doesn't happen on success.
func TestRegisterK8sDqliteReverter_Success(t *testing.T) {
	g := NewWithT(t)

	tmpDir := t.TempDir()
	dqliteStateDir := filepath.Join(tmpDir, "dqlite")
	g.Expect(os.MkdirAll(dqliteStateDir, 0o755)).To(Succeed())

	testFile := filepath.Join(dqliteStateDir, "test.db")
	g.Expect(os.WriteFile(testFile, []byte("test data"), 0o644)).To(Succeed())

	mockSnap := &snapmock.Snap{
		Mock: snapmock.Mock{
			K8sDqliteStateDir: dqliteStateDir,
		},
	}
	reverter := revert.New()
	defer reverter.Fail()

	// Register the reverter
	registerK8sDqliteReverter(mockSnap, reverter)

	// Mark as successful (no cleanup should happen)
	reverter.Success()

	// Verify directory still exists
	g.Expect(testFile).To(BeAnExistingFile())
}

// TestRegisterEtcdMemberReverter_NotEnoughEndpoints tests that cleanup is skipped when <3 endpoints.
func TestRegisterEtcdMemberReverter_NotEnoughEndpoints(t *testing.T) {
	g := NewWithT(t)

	tmpDir := t.TempDir()
	etcdDir := filepath.Join(tmpDir, "etcd")
	g.Expect(os.MkdirAll(etcdDir, 0o755)).To(Succeed())

	testFile := filepath.Join(etcdDir, "member/snap/db")
	g.Expect(os.MkdirAll(filepath.Dir(testFile), 0o755)).To(Succeed())
	g.Expect(os.WriteFile(testFile, []byte("etcd data"), 0o644)).To(Succeed())

	// Only 2 endpoints - RegisterEtcdMemberReverter skips etcd operations when <3
	endpoints := []string{"https://node1:2379", "https://node2:2379"}

	mockSnap := &snapmock.Snap{
		Mock: snapmock.Mock{
			EtcdDir: etcdDir,
			// No EtcdClient needed - reverter won't call snap.EtcdClient with <3 endpoints
		},
	}
	reverter := revert.New()

	registerEtcdMemberReverter(mockSnap, "node2", endpoints, reverter)

	// Trigger reverter
	reverter.Fail()

	// Verify directory was NOT cleaned up (quorum protection)
	g.Expect(etcdDir).To(BeAnExistingFile())
}

// TestRegisterEtcdMemberReverter_ClientCreationFailure tests error handling when EtcdClient creation fails.
func TestRegisterEtcdMemberReverter_ClientCreationFailure(t *testing.T) {
	g := NewWithT(t)

	tmpDir := t.TempDir()
	etcdDir := filepath.Join(tmpDir, "etcd")
	g.Expect(os.MkdirAll(etcdDir, 0o755)).To(Succeed())

	testFile := filepath.Join(etcdDir, "member/snap/db")
	g.Expect(os.MkdirAll(filepath.Dir(testFile), 0o755)).To(Succeed())
	g.Expect(os.WriteFile(testFile, []byte("etcd data"), 0o644)).To(Succeed())
	// 3 endpoints - should attempt etcd operations but client creation fails
	endpoints := []string{"https://node1:2379", "https://node2:2379", "https://node3:2379"}
	nodeName := "node2"

	// Mock snap that returns an error from EtcdClient
	mockSnap := &snapmock.Snap{
		Mock: snapmock.Mock{
			EtcdDir:       etcdDir,
			EtcdClientErr: fmt.Errorf("failed to create etcd client"),
		},
	}
	reverter := revert.New()

	registerEtcdMemberReverter(mockSnap, nodeName, endpoints, reverter)

	// Trigger reverter
	reverter.Fail()

	// Verify directory was NOT cleaned up when client creation fails
	g.Expect(etcdDir).To(BeAnExistingFile(), "etcd directory should NOT be removed when client creation fails")
}

// TestRegisterK8sNodeDeletionReverter_FailDeletesNode ensures a failed join triggers Node deletion.
func TestRegisterK8sNodeDeletionReverter_FailDeletesNode(t *testing.T) {
	g := NewWithT(t)

	nodeName := "test-node"
	clientset := fake.NewClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	})
	k8sClient := &kubernetes.Client{Interface: clientset}

	reverter := revert.New()
	registerK8sNodeDeletionReverter(k8sClient, nodeName, reverter)

	// Simulate join failure
	reverter.Fail()

	// Node should be deleted
	_, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue(), "node should be removed on revert")
}

// TestRegisterK8sNodeDeletionReverter_Success ensures a successful join does not delete the Node.
func TestRegisterK8sNodeDeletionReverter_Success(t *testing.T) {
	g := NewWithT(t)

	nodeName := "test-node"
	clientset := fake.NewClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: nodeName},
	})
	k8sClient := &kubernetes.Client{Interface: clientset}

	reverter := revert.New()
	defer reverter.Fail()

	registerK8sNodeDeletionReverter(k8sClient, nodeName, reverter)

	// Mark as successful join (reverts should not run)
	reverter.Success()

	// Node should still exist
	_, err := clientset.CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
	g.Expect(err).NotTo(HaveOccurred(), "node should remain when join succeeds")
}
