package snaputil_test

import (
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
)

func TestIsWorker(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{
			LockFilesDir: t.TempDir(),
		},
	}

	t.Run("WorkerFileExists", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fname := path.Join(mock.LockFilesDir(), "worker")
		lock, err := os.Create(fname)
		g.Expect(err).ToNot(HaveOccurred())
		lock.Close()

		exists, err := snaputil.IsWorker(mock)
		g.Expect(err).To(BeNil())
		g.Expect(exists).To(BeTrue())
	})

	t.Run("WorkerFileNotExists", func(t *testing.T) {
		mock.Mock.LockFilesDir = "/non-existent"
		g := NewGomegaWithT(t)
		exists, err := snaputil.IsWorker(mock)
		g.Expect(err).To(BeNil())
		g.Expect(exists).To(BeFalse())
	})
}

func TestMarkAsWorkerNode(t *testing.T) {
	lockFilesDir := t.TempDir()
	mock := &mock.Snap{
		Mock: mock.Mock{
			LockFilesDir: lockFilesDir,
			UID:          1000,
			GID:          1000,
		},
	}

	t.Run("MarkWorker", func(t *testing.T) {
		g := NewGomegaWithT(t)
		err := snaputil.MarkAsWorkerNode(mock, true)
		g.Expect(err).To(BeNil())

		workerFile := path.Join(mock.LockFilesDir(), "worker")
		g.Expect(workerFile).To(BeAnExistingFile())

		// Clean up
		err = os.Remove(workerFile)
		g.Expect(err).To(BeNil())
	})

	t.Run("UnmarkWorker", func(t *testing.T) {
		g := NewGomegaWithT(t)
		workerFile := path.Join(mock.LockFilesDir(), "worker")
		_, err := os.Create(workerFile)
		g.Expect(err).To(BeNil())

		err = snaputil.MarkAsWorkerNode(mock, false)
		g.Expect(err).To(BeNil())

		g.Expect(workerFile).NotTo(BeAnExistingFile())
	})
}

func TestMarkAsWorkerNode_ErrorCases(t *testing.T) {
	mock := &mock.Snap{
		Mock: mock.Mock{
			LockFilesDir: t.TempDir(),
			UID:          1000,
			GID:          1000,
		},
	}

	t.Run("FailedToCreateWorkerFile", func(t *testing.T) {
		mock.Mock.LockFilesDir = "/non-existent"
		g := NewGomegaWithT(t)
		err := snaputil.MarkAsWorkerNode(mock, true)
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("FailedToRemoveWorkerFile", func(t *testing.T) {
		mock.Mock.LockFilesDir = "/non-existent"
		g := NewGomegaWithT(t)
		err := snaputil.MarkAsWorkerNode(mock, false)
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("FailedToChownWorkerFile", func(t *testing.T) {
		mock.Mock.UID = -1 // Invalid UID to cause chown failure
		g := NewGomegaWithT(t)
		err := snaputil.MarkAsWorkerNode(mock, true)
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("FailedToChmodWorkerFile", func(t *testing.T) {
		mock.Mock.LockFilesDir = "/non-existent"
		g := NewGomegaWithT(t)
		err := snaputil.MarkAsWorkerNode(mock, true)
		g.Expect(err).To(HaveOccurred())
	})
}
