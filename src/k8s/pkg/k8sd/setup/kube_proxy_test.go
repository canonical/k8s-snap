package setup_test

import (
	"context"
	"os"
	"path"
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
			KubernetesConfigDir: path.Join(dir, "kubernetes"),
			ServiceArgumentsDir: path.Join(dir, "args"),
			OnLXD:               false,
			UID:                 os.Getuid(),
			GID:                 os.Getgid(),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())

	t.Run("Args", func(t *testing.T) {
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16", nil)).To(BeNil())

		for key, expectedVal := range map[string]string{
			"--cluster-cidr":           "10.1.0.0/16",
			"--hostname-override":      "myhostname",
			"--kubeconfig":             path.Join(dir, "kubernetes", "proxy.conf"),
			"--profiling":              "false",
			"--conntrack-max-per-core": "",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", key)
				g.Expect(err).To(BeNil())
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
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16", extraArgs)).To(BeNil())

		for key, expectedVal := range map[string]string{
			"--cluster-cidr":           "10.1.0.0/16",
			"--hostname-override":      "myoverriddenhostname",
			"--kubeconfig":             path.Join(dir, "kubernetes", "proxy.conf"),
			"--profiling":              "false",
			"--conntrack-max-per-core": "",
			"--my-extra-arg":           "my-extra-val",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", key)
				g.Expect(err).To(BeNil())
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
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16", nil)).To(BeNil())

		for key, expectedVal := range map[string]string{
			"--conntrack-max-per-core": "0",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "kube-proxy", key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(expectedVal))
			})
		}
	})

}
