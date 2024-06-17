package setup_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func setK8sDqliteMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir: path.Join(dir, "args"),
		K8sDqliteStateDir:   path.Join(dir, "k8s-dqlite"),
	}
}

func TestK8sDqlite(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setK8sDqliteMock)

		// Call the K8sDqlite setup function with mock arguments
		g.Expect(setup.K8sDqlite(s, "192.168.0.1:1234", []string{"192.168.0.1:1234"}, nil)).To(BeNil())

		// Ensure the K8sDqlite arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--listen", expectedVal: fmt.Sprintf("unix://%s", path.Join(s.Mock.K8sDqliteStateDir, "k8s-dqlite.sock"))},
			{key: "--storage-dir", expectedVal: s.Mock.K8sDqliteStateDir},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-dqlite", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		// Ensure the K8sDqlite arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "k8s-dqlite"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("WithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setK8sDqliteMock)

		extraArgs := map[string]*string{
			"--my-extra-arg": utils.Pointer("my-extra-val"),
			"--listen":       nil,
			"--storage-dir":  utils.Pointer("overridden-storage-dir"),
		}
		// Call the K8sDqlite setup function with mock arguments
		g.Expect(setup.K8sDqlite(s, "192.168.0.1:1234", []string{"192.168.0.1:1234"}, extraArgs)).To(BeNil())

		// Ensure the K8sDqlite arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--storage-dir", expectedVal: "overridden-storage-dir"},
			{key: "--my-extra-arg", expectedVal: "my-extra-val"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-dqlite", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(tc.expectedVal))
			})
		}

		// --listen was deleted by extraArgs
		t.Run("--listen", func(t *testing.T) {
			g := NewWithT(t)
			val, err := snaputil.GetServiceArgument(s, "k8s-dqlite", "--listen")
			g.Expect(err).To(BeNil())
			g.Expect(val).To(BeZero())
		})

		// Ensure the K8sDqlite arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "k8s-dqlite"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("YAMLFileContents", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setK8sDqliteMock)

		expectedYaml := "Address: 192.168.0.1:1234\nCluster:\n- 192.168.0.1:1234\n- 192.168.0.2:1234\n- 192.168.0.3:1234\n"

		cluster := []string{
			"192.168.0.1:1234",
			"192.168.0.2:1234",
			"192.168.0.3:1234",
		}

		g.Expect(setup.K8sDqlite(s, "192.168.0.1:1234", cluster, nil)).To(BeNil())

		b, err := os.ReadFile(path.Join(s.Mock.K8sDqliteStateDir, "init.yaml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(b)).To(Equal(expectedYaml))
	})

	t.Run("MissingStateDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sDqliteMock)

		s.Mock.K8sDqliteStateDir = "nonexistent"

		g.Expect(setup.K8sDqlite(s, "", []string{}, nil)).ToNot(Succeed())
	})

	t.Run("MissingArgsDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sDqliteMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"

		g.Expect(setup.K8sDqlite(s, "", []string{}, nil)).ToNot(Succeed())
	})
}
