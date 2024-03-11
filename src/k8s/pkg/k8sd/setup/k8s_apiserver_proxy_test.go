package setup_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/proxy"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func setK8sApiServerMock(s *mock.Snap, dir string) {
	s.Mock = mock.Mock{
		ServiceArgumentsDir:   path.Join(dir, "args"),
		ServiceExtraConfigDir: path.Join(dir, "args/conf.d"),
	}
}

func TestK8sApiServerProxy(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		g.Expect(setup.K8sAPIServerProxy(s, []string{}))

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
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("MissingExtraConfigDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		s.Mock.ServiceExtraConfigDir = "nonexistent"
		g.Expect(setup.K8sAPIServerProxy(s, nil)).ToNot(Succeed())
	})

	t.Run("MissingServiceArgumentsDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"
		g.Expect(setup.K8sAPIServerProxy(s, nil)).ToNot(Succeed())
	})

	t.Run("JSONFileContent", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		endpoints := []string{"192.168.0.1", "192.168.0.2", "192.168.0.3"}
		fileName := path.Join(s.Mock.ServiceExtraConfigDir, "k8s-apiserver-proxy.json")

		g.Expect(setup.K8sAPIServerProxy(s, endpoints)).To(Succeed())

		b, err := os.ReadFile(fileName)
		g.Expect(err).NotTo(HaveOccurred())

		var config proxy.Configuration
		err = json.Unmarshal(b, &config)
		g.Expect(err).NotTo(HaveOccurred())

		// Compare the expected endpoints with those in the file
		g.Expect(config.Endpoints).To(Equal(endpoints))
	})
}
