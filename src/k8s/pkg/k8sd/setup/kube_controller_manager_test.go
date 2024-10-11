package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func setKubeControllerManagerMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir: filepath.Join(dir, "args"),
		KubernetesConfigDir: filepath.Join(dir, "k8s-config"),
		KubernetesPKIDir:    filepath.Join(dir, "k8s-pki"),
	}
}

func TestKubeControllerManager(t *testing.T) {
	t.Run("WithClusterSigning", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeControllerManagerMock)

		// Create ca.key so that cluster-signing-cert-file and cluster-signing-key-file are added to the arguments
		os.Create(filepath.Join(s.Mock.KubernetesPKIDir, "ca.key"))

		// Call the kube controller manager setup function
		g.Expect(setup.KubeControllerManager(s, nil)).To(Succeed())

		// Ensure the kube controller manager arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--authorization-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--leader-elect-lease-duration", expectedVal: "30s"},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--root-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--service-account-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--terminated-pod-gc-threshold", expectedVal: "12500"},
			{key: "--use-service-account-credentials", expectedVal: "true"},
			{key: "--cluster-signing-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--cluster-signing-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.key")},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kube-controller-manager"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))

		t.Run("MissingArgsDir", func(t *testing.T) {
			g := NewWithT(t)
			s.Mock.ServiceArgumentsDir = "nonexistent"
			g.Expect(setup.KubeControllerManager(s, nil)).ToNot(Succeed())
		})
	})

	t.Run("WithoutClusterSigning", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeControllerManagerMock)

		// Call the kube controller manager setup function
		g.Expect(setup.KubeControllerManager(s, nil)).To(Succeed())

		// Ensure the kube controller manager arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--authorization-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--leader-elect-lease-duration", expectedVal: "30s"},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "false"},
			{key: "--root-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--service-account-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--terminated-pod-gc-threshold", expectedVal: "12500"},
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
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kube-controller-manager"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))

		t.Run("MissingArgsDir", func(t *testing.T) {
			g := NewWithT(t)
			s.Mock.ServiceArgumentsDir = "nonexistent"
			g.Expect(setup.KubeControllerManager(s, nil)).ToNot(Succeed())
		})
	})

	t.Run("WithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeControllerManagerMock)

		// Create ca.key so that cluster-signing-cert-file and cluster-signing-key-file are added to the arguments
		os.Create(filepath.Join(s.Mock.KubernetesPKIDir, "ca.key"))

		extraArgs := map[string]*string{
			"--leader-elect-lease-duration": nil,
			"--profiling":                   utils.Pointer("true"),
			"--my-extra-arg":                utils.Pointer("my-extra-val"),
		}
		// Call the kube controller manager setup function
		g.Expect(setup.KubeControllerManager(s, extraArgs)).To(Succeed())

		// Ensure the kube controller manager arguments file has the expected arguments and values
		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--authentication-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--authorization-kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "controller.conf")},
			{key: "--leader-elect-renew-deadline", expectedVal: "15s"},
			{key: "--profiling", expectedVal: "true"},
			{key: "--root-ca-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--service-account-private-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "serviceaccount.key")},
			{key: "--terminated-pod-gc-threshold", expectedVal: "12500"},
			{key: "--use-service-account-credentials", expectedVal: "true"},
			{key: "--cluster-signing-cert-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.crt")},
			{key: "--cluster-signing-key-file", expectedVal: filepath.Join(s.Mock.KubernetesPKIDir, "ca.key")},
			{key: "--my-extra-arg", expectedVal: "my-extra-val"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-controller-manager", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}

		// Ensure that the leader-elect-lease-duration argument was deleted
		val, err := snaputil.GetServiceArgument(s, "kube-controller-manager", "--leader-elect-lease-duration")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(val).To(BeZero())

		// Ensure the kube controller manager arguments file has exactly the expected number of arguments
		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "kube-controller-manager"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(args).To(HaveLen(len(tests)))

		t.Run("MissingArgsDir", func(t *testing.T) {
			g := NewWithT(t)
			s.Mock.ServiceArgumentsDir = "nonexistent"
			g.Expect(setup.KubeControllerManager(s, nil)).ToNot(Succeed())
		})
	})
}
