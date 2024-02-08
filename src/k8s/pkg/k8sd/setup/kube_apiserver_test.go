package setup_test

import (
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestKubeAPIServer(t *testing.T) {
	g := NewWithT(t)

	dir := "testdata/apiserver"

	s := &mock.Snap{
		Mock: mock.Mock{
			UID:                   os.Getuid(),
			GID:                   os.Getgid(),
			KubernetesConfigDir:   path.Join(dir, "kubernetes"),
			KubernetesPKIDir:      path.Join(dir, "kubernetes-pki"),
			ServiceArgumentsDir:   path.Join(dir, "args"),
			ServiceExtraConfigDir: path.Join(dir, "args/conf.d"),
			K8sDqliteStateDir:     path.Join(dir, "k8s-dqlite"),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())
	g.Expect(setup.KubeAPIServer(s, "10.152.0.0/16", "https://10.0.0.1:6400/1.0/kubernetes/auth/webhook", false, "k8s-dqlite", "Node,RBAC")).To(BeNil())

	// t.Run("Config", func(t *testing.T) {
	// 	g := NewWithT(t)
	// 	b, err := os.ReadFile(path.Join(dir, "containerd", "config.toml"))
	// 	g.Expect(err).To(BeNil())
	// 	g.Expect(string(b)).To(SatisfyAll(
	// 		ContainSubstring(fmt.Sprintf(`imports = ["%s/*.toml"]`, path.Join(dir, "containerd-confd"))),
	// 		ContainSubstring(fmt.Sprintf(`conf_dir = "%s"`, path.Join(dir, "cni-netd"))),
	// 		ContainSubstring(fmt.Sprintf(`bin_dir = "%s"`, path.Join(dir, "opt-cni-bin"))),
	// 		ContainSubstring(fmt.Sprintf(`config_path = "%s"`, path.Join(dir, "containerd-registries"))),
	// 	))

	// 	info, err := os.Stat(path.Join(dir, "containerd", "config.toml"))
	// 	g.Expect(err).To(BeNil())
	// 	g.Expect(info.Mode().Perm()).To(Equal(fs.FileMode(0600)))

	// 	switch stat := info.Sys().(type) {
	// 	case *syscall.Stat_t:
	// 		g.Expect(stat.Uid).To(Equal(uint32(os.Getuid())))
	// 		g.Expect(stat.Gid).To(Equal(uint32(os.Getuid())))
	// 	default:
	// 		g.Fail("failed to stat config.toml")
	// 	}
	// })

	// t.Run("CNI", func(t *testing.T) {
	// 	g := NewWithT(t)
	// 	for _, plugin := range []string{"plugin1", "plugin2"} {
	// 		link, err := os.Readlink(path.Join(dir, "opt-cni-bin", plugin))
	// 		g.Expect(err).To(BeNil())
	// 		g.Expect(link).To(Equal("cni"))
	// 	}

	// 	info, err := os.Stat(path.Join(dir, "opt-cni-bin"))
	// 	g.Expect(err).To(BeNil())
	// 	g.Expect(info.Mode().Perm()).To(Equal(fs.FileMode(0700)))

	// 	switch stat := info.Sys().(type) {
	// 	case *syscall.Stat_t:
	// 		g.Expect(stat.Uid).To(Equal(uint32(os.Getuid())))
	// 		g.Expect(stat.Gid).To(Equal(uint32(os.Getuid())))
	// 	default:
	// 		g.Fail("failed to stat installed cni")
	// 	}
	// })

	// t.Run("Args", func(t *testing.T) {
	// 	for key, expectedVal := range map[string]string{
	// 		"--config":  path.Join(dir, "containerd", "config.toml"),
	// 		"--state":   path.Join(dir, "containerd-state"),
	// 		"--root":    path.Join(dir, "containerd-root"),
	// 		"--address": path.Join(dir, "containerd-run", "containerd.sock"),
	// 	} {
	// 		t.Run(key, func(t *testing.T) {
	// 			g := NewWithT(t)
	// 			val, err := snaputil.GetServiceArgument(s, "containerd", key)
	// 			g.Expect(err).To(BeNil())
	// 			g.Expect(val).To(Equal(expectedVal))
	// 		})
	// 	}
	// })

}
