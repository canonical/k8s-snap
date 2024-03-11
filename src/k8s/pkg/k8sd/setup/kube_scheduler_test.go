package setup_test

import (
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func setKubeSchedulerMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir: path.Join(dir, "args"),
		KubernetesConfigDir: path.Join(dir, "k8s-config"),
	}
}

func TestKubeScheduler(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeSchedulerMock)

		// Call the kube scheduler setup function
		g.Expect(setup.KubeScheduler(s)).To(BeNil())

		// Ensure the kube scheduler arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--authorization-kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--leader-elect-lease-duration", expectedVal: "30s"},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "false"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-scheduler", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kube scheduler arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-scheduler"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))

	})

	t.Run("MissingArgsDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setKubeSchedulerMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"

		g.Expect(setup.KubeScheduler(s)).ToNot(Succeed())
	})
}
