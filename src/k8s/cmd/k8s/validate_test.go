package k8s

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

func TestValidateConfig(t *testing.T) {
	t.Run("Bootstrap", func(t *testing.T) {
		g := NewWithT(t)
		// Replace the ContainerdSocketDir to avoid checking against a real containerd.sock that may be running.
		containerdBaseDir, err := os.MkdirTemp("", "test-containerd")
		g.Expect(err).To(Not(HaveOccurred()))
		defer os.RemoveAll(containerdBaseDir)

		bootstrapConfig := apiv1.BootstrapConfig{
			ContainerdBaseDir: containerdBaseDir,
		}

		mockRunner := &mock.Runner{}
		err = verifyBootstrapConfigWithRunCommand(bootstrapConfig, mockRunner.Run)
		g.Expect(err).To(Not(HaveOccurred()))

		t.Run("Fail port already in use", func(t *testing.T) {
			g := NewWithT(t)
			// Open a port which will be checked (kubelet).
			port := "9999"
			bootstrapConfig := apiv1.BootstrapConfig{
				ContainerdBaseDir:    containerdBaseDir,
				ExtraNodeKubeletArgs: map[string]*string{"--port": &port},
			}

			l, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			g.Expect(err).To(Not(HaveOccurred()))
			defer l.Close()

			err = verifyBootstrapConfigWithRunCommand(bootstrapConfig, mockRunner.Run)
			g.Expect(err).To(HaveOccurred())
		})

		t.Run("Fail socket exists", func(t *testing.T) {
			g := NewWithT(t)

			// Create the containerd.sock file, which should cause the check to fail.
			containerdDir := filepath.Join(containerdBaseDir, "run", "containerd")
			err := os.MkdirAll(containerdDir, os.ModeDir)
			g.Expect(err).To(Not(HaveOccurred()))

			f, err := os.Create(filepath.Join(containerdDir, "containerd.sock"))
			g.Expect(err).To(Not(HaveOccurred()))
			f.Close()
			defer os.Remove(f.Name())

			err = verifyBootstrapConfigWithRunCommand(bootstrapConfig, mockRunner.Run)
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Join", func(t *testing.T) {
		g := NewWithT(t)
		// Replace the ContainerdSocketDir to avoid checking against a real containerd.sock that may be running.
		containerdBaseDir, err := os.MkdirTemp("", "test-containerd")
		g.Expect(err).To(Not(HaveOccurred()))
		defer os.RemoveAll(containerdBaseDir)

		mockRunner := &mock.Runner{}

		t.Run("worker node", func(t *testing.T) {
			g := NewWithT(t)

			joinConfig := &apiv1.WorkerJoinConfig{
				ContainerdBaseDir: containerdBaseDir,
			}
			joinConfigBytes, err := yaml.Marshal(joinConfig)
			g.Expect(err).To(Not(HaveOccurred()))

			internalToken := types.InternalWorkerNodeToken{
				Token:  "foo",
				Secret: "lish",
			}
			tokenString, err := internalToken.Encode()
			g.Expect(err).To(Not(HaveOccurred()))

			err = verifyJoinConfigWithRunCommand(string(joinConfigBytes), tokenString, mockRunner.Run)
			g.Expect(err).To(Not(HaveOccurred()))
		})

		t.Run("control plane node", func(t *testing.T) {
			g := NewWithT(t)

			joinConfig := &apiv1.ControlPlaneJoinConfig{
				ContainerdBaseDir: containerdBaseDir,
			}
			joinConfigBytes, err := yaml.Marshal(joinConfig)
			g.Expect(err).To(Not(HaveOccurred()))

			err = verifyJoinConfigWithRunCommand(string(joinConfigBytes), "", mockRunner.Run)
			g.Expect(err).To(Not(HaveOccurred()))
		})
	})
}
