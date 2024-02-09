package setup_test

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
)

func TestContainerd(t *testing.T) {
	g := NewWithT(t)

	dir := t.TempDir()

	g.Expect(os.WriteFile(path.Join(dir, "mockcni"), []byte("echo hi"), 0600)).To(BeNil())

	s := &mock.Snap{
		Mock: mock.Mock{
			ContainerdConfigDir:         path.Join(dir, "containerd"),
			ContainerdRootDir:           path.Join(dir, "containerd-root"),
			ContainerdSocketDir:         path.Join(dir, "containerd-run"),
			ContainerdRegistryConfigDir: path.Join(dir, "containerd-registries"),
			ContainerdStateDir:          path.Join(dir, "containerd-state"),
			ContainerdExtraConfigDir:    path.Join(dir, "containerd-confd"),
			ServiceArgumentsDir:         path.Join(dir, "args"),
			CNIBinDir:                   path.Join(dir, "opt-cni-bin"),
			CNIConfDir:                  path.Join(dir, "cni-netd"),
			CNIPluginsBinary:            path.Join(dir, "mockcni"),
			CNIPlugins:                  []string{"plugin1", "plugin2"},
			UID:                         os.Getuid(),
			GID:                         os.Getgid(),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())
	g.Expect(setup.Containerd(s)).To(BeNil())

	t.Run("Config", func(t *testing.T) {
		g := NewWithT(t)
		b, err := os.ReadFile(path.Join(dir, "containerd", "config.toml"))
		g.Expect(err).To(BeNil())
		g.Expect(string(b)).To(SatisfyAll(
			ContainSubstring(fmt.Sprintf(`imports = ["%s/*.toml"]`, path.Join(dir, "containerd-confd"))),
			ContainSubstring(fmt.Sprintf(`conf_dir = "%s"`, path.Join(dir, "cni-netd"))),
			ContainSubstring(fmt.Sprintf(`bin_dir = "%s"`, path.Join(dir, "opt-cni-bin"))),
			ContainSubstring(fmt.Sprintf(`config_path = "%s"`, path.Join(dir, "containerd-registries"))),
		))

		info, err := os.Stat(path.Join(dir, "containerd", "config.toml"))
		g.Expect(err).To(BeNil())
		g.Expect(info.Mode().Perm()).To(Equal(fs.FileMode(0600)))

		switch stat := info.Sys().(type) {
		case *syscall.Stat_t:
			g.Expect(stat.Uid).To(Equal(uint32(os.Getuid())))
			g.Expect(stat.Gid).To(Equal(uint32(os.Getgid())))
		default:
			g.Fail("failed to stat config.toml")
		}
	})

	t.Run("CNI", func(t *testing.T) {
		g := NewWithT(t)
		for _, plugin := range []string{"plugin1", "plugin2"} {
			link, err := os.Readlink(path.Join(dir, "opt-cni-bin", plugin))
			g.Expect(err).To(BeNil())
			g.Expect(link).To(Equal("cni"))
		}

		info, err := os.Stat(path.Join(dir, "opt-cni-bin"))
		g.Expect(err).To(BeNil())
		g.Expect(info.Mode().Perm()).To(Equal(fs.FileMode(0700)))

		switch stat := info.Sys().(type) {
		case *syscall.Stat_t:
			g.Expect(stat.Uid).To(Equal(uint32(os.Getuid())))
			g.Expect(stat.Gid).To(Equal(uint32(os.Getgid())))
		default:
			g.Fail("failed to stat installed cni")
		}
	})

	t.Run("Args", func(t *testing.T) {
		for key, expectedVal := range map[string]string{
			"--config":  path.Join(dir, "containerd", "config.toml"),
			"--state":   path.Join(dir, "containerd-state"),
			"--root":    path.Join(dir, "containerd-root"),
			"--address": path.Join(dir, "containerd-run", "containerd.sock"),
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "containerd", key)
				g.Expect(err).To(BeNil())
				g.Expect(val).To(Equal(expectedVal))
			})
		}
	})

}
