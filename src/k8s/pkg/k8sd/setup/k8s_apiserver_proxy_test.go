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

func mustReturnK8sApiServerMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir:   path.Join(dir, "args"),
		ServiceExtraConfigDir: path.Join(dir, "args/conf.d"),
	}
}

func TestK8sApiServerProxy(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		s, _ := mustSetupSnapAndDirectories(t, mustReturnK8sApiServerMock)

		g.Expect(setup.K8sAPIServerProxy(s, []string{}))

		os.Create(path.Join(s.Mock.ServiceExtraConfigDir, "k8s-apiserver-proxy.json"))

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--endpoints", expectedVal: path.Join(s.Mock.ServiceExtraConfigDir, "k8s-apiserver-proxy.json")},
			{key: "--kubeconfig", expectedVal: path.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
			{key: "--listen", expectedVal: "127.0.0.1:6443"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-apiserver-proxy", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}
		args, err := utils.ParseArgumentFile(path.Join(s.Mock.ServiceArgumentsDir, "k8s-apiserver-proxy"))
		g.Expect(err).To(BeNil())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("MissingExtraConfigDir", func(t *testing.T) {
		g := NewWithT(t)

		s, _ := mustSetupSnapAndDirectories(t, mustReturnK8sApiServerMock)

		s.Mock.ServiceExtraConfigDir = "nonexistent"
		g.Expect(setup.K8sAPIServerProxy(s, []string{})).ToNot(Succeed())
	})

	t.Run("MissingServiceArgumentsDir", func(t *testing.T) {
		g := NewWithT(t)

		s, _ := mustSetupSnapAndDirectories(t, mustReturnK8sApiServerMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"
		g.Expect(setup.K8sAPIServerProxy(s, []string{})).ToNot(Succeed())
	})
}
