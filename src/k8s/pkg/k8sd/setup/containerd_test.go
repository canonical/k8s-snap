package setup_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestContainerd(t *testing.T) {
	g := NewWithT(t)

	dir := t.TempDir()

	g.Expect(utils.WriteFile(filepath.Join(dir, "mockcni"), []byte("echo hi"), 0o600)).To(Succeed())

	s := &mock.Snap{
		Mock: mock.Mock{
			ContainerdConfigDir:         filepath.Join(dir, "containerd"),
			ContainerdRootDir:           filepath.Join(dir, "containerd-root"),
			ContainerdSocketDir:         filepath.Join(dir, "containerd-run"),
			ContainerdSocketPath:        filepath.Join(dir, "containerd-run", "containerd.sock"),
			ContainerdRegistryConfigDir: filepath.Join(dir, "containerd-hosts"),
			ContainerdStateDir:          filepath.Join(dir, "containerd-state"),
			ContainerdExtraConfigDir:    filepath.Join(dir, "containerd-confd"),
			ServiceArgumentsDir:         filepath.Join(dir, "args"),
			CNIBinDir:                   filepath.Join(dir, "opt-cni-bin"),
			CNIConfDir:                  filepath.Join(dir, "cni-netd"),
			CNIPluginsBinary:            filepath.Join(dir, "mockcni"),
			CNIPlugins:                  []string{"plugin1", "plugin2"},
			UID:                         os.Getuid(),
			GID:                         os.Getgid(),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())
	g.Expect(setup.Containerd(s, map[string]any{
		"imports": []string{"/custom/imports/*.toml"},
	}, map[string]*string{
		"--log-level":    utils.Pointer("debug"),
		"--metrics":      utils.Pointer("true"),
		"--address":      nil, // This should trigger a delete
		"--my-extra-arg": utils.Pointer("my-extra-val"),
	})).To(Succeed())

	t.Run("Config", func(t *testing.T) {
		g := NewWithT(t)
		b, err := os.ReadFile(filepath.Join(dir, "containerd", "config.toml"))
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(string(b)).To(SatisfyAll(
			ContainSubstring(fmt.Sprintf(`imports = ["%s/*.toml", "/custom/imports/*.toml"]`, filepath.Join(dir, "containerd-confd"))),
			ContainSubstring(fmt.Sprintf(`conf_dir = "%s"`, filepath.Join(dir, "cni-netd"))),
			ContainSubstring(fmt.Sprintf(`bin_dir = "%s"`, filepath.Join(dir, "opt-cni-bin"))),
			ContainSubstring(fmt.Sprintf(`config_path = "%s"`, filepath.Join(dir, "containerd-hosts"))),
		))

		info, err := os.Stat(filepath.Join(dir, "containerd", "config.toml"))
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(info.Mode().Perm()).To(Equal(fs.FileMode(0o600)))

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
			link, err := os.Readlink(filepath.Join(dir, "opt-cni-bin", plugin))
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(link).To(Equal("cni"))
		}

		info, err := os.Stat(filepath.Join(dir, "opt-cni-bin"))
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(info.Mode().Perm()).To(Equal(fs.FileMode(0o700)))

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
			"--config":       filepath.Join(dir, "containerd", "config.toml"),
			"--root":         filepath.Join(dir, "containerd-root"),
			"--state":        filepath.Join(dir, "containerd-state"),
			"--log-level":    "debug",
			"--metrics":      "true",
			"--my-extra-arg": "my-extra-val",
		} {
			t.Run(key, func(t *testing.T) {
				g := NewWithT(t)
				val, err := snaputil.GetServiceArgument(s, "containerd", key)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(val).To(Equal(expectedVal))
			})
		}
		// --address was deleted by extraArgs
		t.Run("--address", func(t *testing.T) {
			g := NewWithT(t)
			val, err := snaputil.GetServiceArgument(s, "containerd", "--address")
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(val).To(BeZero())
		})
	})
}
