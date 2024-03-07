package setup_test

import (
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func mustReturnMockForKubeControllerManager(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir: path.Join(dir, "args"),
		KubernetesConfigDir: path.Join(dir, "k8s-config"),
		KubernetesPKIDir:    path.Join(dir, "k8s-pki"),
	}
}

func TestKubeControllerManager(t *testing.T) {
	t.Run("ArgsWithClusterSigning", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubeControllerManager)

		// Create ca.key so that cluster-signing-cert-file and cluster-signing-key-file are added to the arguments
		os.Create(path.Join(s.Mock.KubernetesPKIDir, "ca.key"))

		// Call the kube controller manager setup function
		g.Expect(setup.KubeControllerManager(s)).To(BeNil())

		// Ensure the kube controller manager arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--authorization-kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--leader-elect-lease-duration", expectedVal: "30s"},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--root-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--service-account-private-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--use-service-account-credentials", expectedVal: "true"},
			{key: "--cluster-signing-cert-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--cluster-signing-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.key")},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-controller-manager", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kube controller manager arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-controller-manager"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))

		t.Run("Error when service arguments dir does not exist", func(t *testing.T) {
			g := NewWithT(t)
			s.Mock.ServiceArgumentsDir = "nonexistent"
			g.Expect(setup.KubeControllerManager(s)).ToNot(Succeed())
		})
	})

	t.Run("ArgsNoClusterSigning", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s, _ := mustSetupSnapAndDirectories(t, mustReturnMockForKubeControllerManager)

		// Call the kube controller manager setup function
		g.Expect(setup.KubeControllerManager(s)).To(BeNil())

		// Ensure the kube controller manager arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--authorization-kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--leader-elect-lease-duration", expectedVal: "30s"},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--root-ca-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--service-account-private-key-file", expectedVal: path.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--use-service-account-credentials", expectedVal: "true"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-controller-manager", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure the kube controller manager arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "kube-controller-manager"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))

		t.Run("MissingArgsDir", func(t *testing.T) {
			g := NewWithT(t)
			s.Mock.ServiceArgumentsDir = "nonexistent"
			g.Expect(setup.KubeControllerManager(s)).ToNot(Succeed())
		})
	})
}
