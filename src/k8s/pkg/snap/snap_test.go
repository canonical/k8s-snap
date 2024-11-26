package snap_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestSnap(t *testing.T) {
	t.Run("NewSnap", func(t *testing.T) {
		t.Run("containerd path with opt", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner := &mock.Runner{}
			snap := snap.NewSnap(snap.SnapOpts{
				RunCommand:        mockRunner.Run,
				ContainerdBaseDir: "/foo/lish",
			})

			g.Expect(snap.ContainerdSocketPath()).To(Equal(filepath.Join("/foo/lish", "run", "containerd", "containerd.sock")))
		})

		t.Run("containerd path classic", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner := &mock.Runner{}
			snap := snap.NewSnap(snap.SnapOpts{
				RunCommand: mockRunner.Run,
			})

			g.Expect(snap.ContainerdSocketPath()).To(Equal(filepath.Join("/", "run", "containerd", "containerd.sock")))
		})

		t.Run("containerd path strict", func(t *testing.T) {
			g := NewWithT(t)
			// We're checking if the snap is strict in NewSnap, which checks the snap.yaml file.
			tmpDir, err := os.MkdirTemp("", "test-snap-k8s")
			g.Expect(err).To(Not(HaveOccurred()))
			defer os.RemoveAll(tmpDir)

			err = os.MkdirAll(filepath.Join(tmpDir, "meta"), os.ModeDir)
			g.Expect(err).To(Not(HaveOccurred()))

			snapData := []byte("confinement: strict\n")
			err = os.WriteFile(filepath.Join(tmpDir, "meta", "snap.yaml"), snapData, 0o644)
			g.Expect(err).To(Not(HaveOccurred()))

			mockRunner := &mock.Runner{}
			snap := snap.NewSnap(snap.SnapOpts{
				SnapDir:       tmpDir,
				SnapCommonDir: tmpDir,
				RunCommand:    mockRunner.Run,
			})

			g.Expect(snap.ContainerdSocketPath()).To(Equal(filepath.Join(tmpDir, "run", "containerd", "containerd.sock")))
		})
	})

	t.Run("Start", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.StartService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl start --enable k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Stop", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})
		err := snap.StopService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl stop --disable k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "test-service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Restart", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.RestartService(context.Background(), "test-service")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("snapctl restart k8s.test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartService(context.Background(), "service")
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("PreInitChecks", func(t *testing.T) {
		g := NewWithT(t)
		// Replace the ContainerdSocketDir to avoid checking against a real containerd.sock that may be running.
		containerdDir, err := os.MkdirTemp("", "test-containerd")
		g.Expect(err).To(Not(HaveOccurred()))
		defer os.RemoveAll(containerdDir)

		mockRunner := &mock.Runner{}
		snap := snap.NewSnap(snap.SnapOpts{
			SnapDir:           "testdir",
			SnapCommonDir:     "testdir",
			RunCommand:        mockRunner.Run,
			ContainerdBaseDir: containerdDir,
		})
		conf := types.ClusterConfig{}

		err = snap.PreInitChecks(context.Background(), conf, true)
		g.Expect(err).To(Not(HaveOccurred()))
		expectedCalls := []string{}
		for _, binary := range []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler", "kube-proxy", "kubelet"} {
			expectedCalls = append(expectedCalls, fmt.Sprintf("testdir/bin/%s --version", binary))
		}
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf(expectedCalls))

		t.Run("Fail port already in use", func(t *testing.T) {
			g := NewWithT(t)
			// Open a port which will be checked (kubelet).
			l, err := net.Listen("tcp", ":10250")
			g.Expect(err).To(Not(HaveOccurred()))
			defer l.Close()

			err = snap.PreInitChecks(context.Background(), conf, true)
			g.Expect(err).To(HaveOccurred())
		})

		t.Run("Fail socket exists", func(t *testing.T) {
			g := NewWithT(t)
			// Create the containerd.sock file, which should cause the check to fail.
			err := os.MkdirAll(snap.ContainerdSocketDir(), os.ModeDir)
			g.Expect(err).To(Not(HaveOccurred()))
			f, err := os.Create(snap.ContainerdSocketPath())
			g.Expect(err).To(Not(HaveOccurred()))
			f.Close()
			defer os.Remove(f.Name())

			err = snap.PreInitChecks(context.Background(), conf, true)
			g.Expect(err).To(HaveOccurred())
		})

		t.Run("Fail run command", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.PreInitChecks(context.Background(), conf, true)
			g.Expect(err).To(HaveOccurred())
		})
	})
}
