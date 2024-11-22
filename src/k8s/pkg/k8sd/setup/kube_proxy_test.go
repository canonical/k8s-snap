package setup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestKubeProxy(t *testing.T) {
	g := NewWithT(t)

	dir := t.TempDir()

	s := &mock.Snap{
		Mock: mock.Mock{
			KubernetesConfigDir: filepath.Join(dir, "kubernetes"),
			ServiceArgumentsDir: filepath.Join(dir, "args"),
			OnLXD:               false,
			UID:                 os.Getuid(),
			GID:                 os.Getgid(),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

	t.Run("Args", func(t *testing.T) {
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16", "127.0.0.1", nil)).To(Succeed())

		for key, expectedVal := range map[string]string{
			"--cluster-cidr":           "10.1.0.0/16",
			"--hostname-override":      "myhostname",
			"--kubeconfig":             filepath.Join(dir, "kubernetes", "proxy.conf"),
			"--profiling":              "false",
			"--conntrack-max-per-core": "",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", key)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(val).To(Equal(expectedVal))
			})
		}
	})

	t.Run("WithExtraArgs", func(t *testing.T) {
		extraArgs := map[string]*string{
			"--hostname-override":    utils.Pointer("myoverriddenhostname"),
			"--healthz-bind-address": nil,
			"--my-extra-arg":         utils.Pointer("my-extra-val"),
		}
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16", "127.0.0.1", extraArgs)).To(Not(HaveOccurred()))

		for key, expectedVal := range map[string]string{
			"--cluster-cidr":           "10.1.0.0/16",
			"--hostname-override":      "myoverriddenhostname",
			"--kubeconfig":             filepath.Join(dir, "kubernetes", "proxy.conf"),
			"--profiling":              "false",
			"--conntrack-max-per-core": "",
			"--my-extra-arg":           "my-extra-val",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", key)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(val).To(Equal(expectedVal))
			})
		}
		// Ensure that the healthz-bind-address argument was deleted
		val, err := snaputil.GetServiceArgument(s, "kube-proxy", "--healthz-bind-address")
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(val).To(BeZero())
	})

	s.Mock.OnLXD = true
	t.Run("ArgsOnLXD", func(t *testing.T) {
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16", "127.0.0.1", nil)).To(Succeed())

		for key, expectedVal := range map[string]string{
			"--conntrack-max-per-core": "0",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", key)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(val).To(Equal(expectedVal))
			})
		}
	})

	t.Run("HostnameOverride", func(t *testing.T) {
		g := NewWithT(t)

		// FIXME(neoaggelos): kube-proxy tests should not reuse the same snap instance, as it leads
		// to implicit state like this shared between the tests
		s.Mock.Hostname = "dev"
		s.Mock.ServiceArgumentsDir = filepath.Join(dir, "k8s")

		g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())
		g.Expect(setup.KubeProxy(context.Background(), s, "dev", "10.1.0.0/16", "127.0.0.1", nil)).To(Succeed())

		val, err := snaputil.GetServiceArgument(s, "kube-proxy", "--hostname-override")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(val).To(BeEmpty())
	})

	t.Run("IPv6", func(t *testing.T) {
		g := NewWithT(t)

		// Create a mock snap
		s := mustSetupSnapAndDirectories(t, setKubeletMock)
		s.Mock.Hostname = "dev"

		g.Expect(setup.KubeProxy(context.Background(), s, "dev", "fd98::/108", "[::1]", nil)).To(Succeed())

		tests := []struct {
			key         string
			expectedVal string
		}{
			{key: "--cluster-cidr", expectedVal: "fd98::/108"},
			{key: "--healthz-bind-address", expectedVal: "[::1]:10256"},
		}
		for _, tc := range tests {
			t.Run(tc.key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", tc.key)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(tc.expectedVal).To(Equal(val))
			})
		}
	})
}
