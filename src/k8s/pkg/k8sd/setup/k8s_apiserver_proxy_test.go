package setup_test

import (
	"encoding/json"
	"os"
	"path/filepath"
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
		ServiceArgumentsDir:   filepath.Join(dir, "args"),
		ServiceExtraConfigDir: filepath.Join(dir, "args/conf.d"),
	}
}

func TestK8sApiServerProxy(t *testing.T) {
	t.Run("Args", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		g.Expect(setup.K8sAPIServerProxy(s, nil, "127.0.0.1", nil)).To(Succeed())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--endpoints", expectedVal: filepath.Join(s.Mock.ServiceExtraConfigDir, "k8s-apiserver-proxy.json")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "kubelet.conf")},
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

		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "k8s-apiserver-proxy"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("WithExtraArgs", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		extraArgs := map[string]*string{
			"--kubeconfig":   utils.Pointer("overridden-kubelet.conf"),
			"--listen":       nil, // This should trigger a delete
			"--my-extra-arg": utils.Pointer("my-extra-val"),
		}
		g.Expect(setup.K8sAPIServerProxy(s, nil, "127.0.0.1", extraArgs)).To(Succeed())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--endpoints", expectedVal: filepath.Join(s.Mock.ServiceExtraConfigDir, "k8s-apiserver-proxy.json")},
			{key: "--kubeconfig", expectedVal: filepath.Join(s.Mock.KubernetesConfigDir, "overridden-kubelet.conf")},
			{key: "--my-extra-arg", expectedVal: "my-extra-val"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-apiserver-proxy", tc.key)
				g.Expect(err).To(BeNil())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}
		// --listen was deleted by extraArgs
		t.Run("--listen", func(t *testing.T) {
			g := NewWithT(t)
			val, err := snaputil.GetServiceArgument(s, "k8s-apiserver-proxy", "--listen")
			g.Expect(err).To(BeNil())
			g.Expect(val).To(BeZero())
		})

		args, err := utils.ParseArgumentFile(filepath.Join(s.Mock.ServiceArgumentsDir, "k8s-apiserver-proxy"))
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(len(args)).To(Equal(len(tests)))
	})

	t.Run("MissingExtraConfigDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		s.Mock.ServiceExtraConfigDir = "nonexistent"
		g.Expect(setup.K8sAPIServerProxy(s, nil, "127.0.0.1", nil)).ToNot(Succeed())
	})

	t.Run("MissingServiceArgumentsDir", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		s.Mock.ServiceArgumentsDir = "nonexistent"
		g.Expect(setup.K8sAPIServerProxy(s, nil, "127.0.0.1", nil)).ToNot(Succeed())
	})

	t.Run("JSONFileContent", func(t *testing.T) {
		g := NewWithT(t)

		s := mustSetupSnapAndDirectories(t, setK8sApiServerMock)

		endpoints := []string{"192.168.0.1", "192.168.0.2", "192.168.0.3"}
		fileName := filepath.Join(s.Mock.ServiceExtraConfigDir, "k8s-apiserver-proxy.json")

		g.Expect(setup.K8sAPIServerProxy(s, endpoints, "127.0.0.1", nil)).To(Succeed())

		b, err := os.ReadFile(fileName)
		g.Expect(err).NotTo(HaveOccurred())

		var config proxy.Configuration
		err = json.Unmarshal(b, &config)
		g.Expect(err).NotTo(HaveOccurred())

		// Compare the expected endpoints with those in the file
		g.Expect(config.Endpoints).To(Equal(endpoints))
	})

	t.Run("IPv6", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)
		s.Mock.Hostname = "dev"

		// Call the kubelet control plane setup function
		g.Expect(setup.K8sAPIServerProxy(s, nil, "[2001:db8::]", nil)).To(Succeed())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--listen", expectedVal: "[2001:db8::]:6443"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "k8s-apiserver-proxy", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}
	})
}
