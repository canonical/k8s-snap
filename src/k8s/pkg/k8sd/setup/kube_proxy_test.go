package setup_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
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
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16")).To(BeNil())

		for key, expectedVal := range map[string]string{
			"--cluster-cidr":           "10.1.0.0/16",
			"--healthz-bind-address":   "127.0.0.1",
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

	s.Mock.OnLXD = true
	t.Run("ArgsOnLXD", func(t *testing.T) {
		g.Expect(setup.KubeProxy(context.Background(), s, "myhostname", "10.1.0.0/16")).To(BeNil())

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
