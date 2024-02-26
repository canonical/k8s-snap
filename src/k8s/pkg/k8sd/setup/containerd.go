package setup

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
)

var (
	containerdConfigTomlTemplate = mustTemplate("containerd", "config.toml")
)

type containerdConfigTomlConfig struct {
	CNIConfDir        string
	CNIBinDir         string
	ImportsDir        string
	RegistryConfigDir string
	PauseImage        string
}

// Containerd configures configuration and arguments for containerd on the local node.
func Containerd(snap snap.Snap) error {
	configToml, err := os.OpenFile(path.Join(snap.ContainerdConfigDir(), "config.toml"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open config.toml: %w", err)
	}
	defer configToml.Close()
	if err := containerdConfigTomlTemplate.Execute(configToml, containerdConfigTomlConfig{
		CNIConfDir:        snap.CNIConfDir(),
		CNIBinDir:         snap.CNIBinDir(),
		ImportsDir:        snap.ContainerdExtraConfigDir(),
		RegistryConfigDir: snap.ContainerdRegistryConfigDir(),
		PauseImage:        "registry.k8s.io/pause:3.7",
	}); err != nil {
		return fmt.Errorf("failed to write config.toml: %w", err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "containerd", map[string]string{
		"--address": path.Join(snap.ContainerdSocketDir(), "containerd.sock"),
		"--config":  path.Join(snap.ContainerdConfigDir(), "config.toml"),
		"--root":    snap.ContainerdRootDir(),
		"--state":   snap.ContainerdStateDir(),
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	cniBinary := path.Join(snap.CNIBinDir(), "cni")
	if err := utils.CopyFile(snap.CNIPluginsBinary(), cniBinary); err != nil {
		return fmt.Errorf("failed to copy cni plugin binary: %w", err)
	}
	if err := os.Chmod(cniBinary, 0700); err != nil {
		return fmt.Errorf("failed to chmod cni plugin binary: %w", err)
	}
	if err := os.Chown(cniBinary, snap.UID(), snap.GID()); err != nil {
		return fmt.Errorf("failed to chown cni plugin binary: %w", err)
	}

	// for each of the CNI plugins, ensure they are a symlink to the "cni" binary
	for _, plugin := range snap.CNIPlugins() {
		pluginInstallPath := path.Join(snap.CNIBinDir(), plugin)

		// if the destination file is already a symlink to "cni", we don't have to do anything
		// if not, then attempt to remove the existing file
		if _, err := os.Stat(pluginInstallPath); err == nil {
			if link, err := os.Readlink(pluginInstallPath); err == nil && link == "cni" {
				continue
			}
			if err := os.Remove(pluginInstallPath); err != nil {
				return fmt.Errorf("failed to remove already existing file %s: %w", pluginInstallPath, err)
			}
		}

		// add plugin as a symlink for the "cni" binary
		if err := os.Symlink("cni", pluginInstallPath); err != nil {
			return fmt.Errorf("failed to symlink cni plugin %s: %w", plugin, err)
		}
	}

	return nil
}
