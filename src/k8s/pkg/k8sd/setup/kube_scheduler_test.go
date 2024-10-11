package setup_test

import (
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func setKubeSchedulerMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir: filepath.Join(dir, "args"),
		KubernetesConfigDir: filepath.Join(dir, "k8s-config"),
	}
}

func TestKubeScheduler(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeSchedulerMock)

		// Call the kube scheduler setup function
		g.Expect(setup.KubeScheduler(s, nil)).To(Succeed())

		// Ensure the kube scheduler arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--authorization-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kube-scheduler"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))

	})

	t.Run("WithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeSchedulerMock)

		extraArgs := map[string]*string{
			"--leader-elect-lease-duration": nil,
			"--profiling":                   utils.Pointer("true"),
			"--my-extra-arg":                utils.Pointer("my-extra-val"),
		}
		// Call the kube scheduler setup function
		g.Expect(setup.KubeScheduler(s, extraArgs)).To(Succeed())

		// Ensure the kube scheduler arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--authorization-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "scheduler.conf")},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "true"},
			{key: "--my-extra-arg", expectedVal: "my-extra-val"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-scheduler", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure that the leader-elect-lease-duration argument was deleted
		val, err := snaputil.GetServiceArgument(s, "kube-scheduler", "--leader-elect-lease-duration")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(val).To(BeZero())

		// Ensure the kube scheduler arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kube-scheduler"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))

	})

	t.Run("MissingArgsDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setKubeSchedulerMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"

		g.Expect(setup.KubeScheduler(s, nil)).ToNot(Succeed())
	})
}
