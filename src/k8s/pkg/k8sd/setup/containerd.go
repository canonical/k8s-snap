package setup

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/pelletier/go-toml"
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

type containerdConfig struct {
	Version int                     `toml:"version"`
	Plugins containerdConfigPlugins `toml:"plugins,omitempty"`
}

type containerdConfigPlugins struct {
	CRI containerdConfigPluginsCRI `toml:"io.containerd.grpc.v1.cri,omitempty"`
}

type containerdConfigPluginsCRI struct {
	Registry containerdConfigPluginsCRIRegistry `toml:"registry,omitempty"`
}

type containerdConfigPluginsCRIRegistry struct {
	Configs map[string]containerdConfigPluginsCRIRegistryConfig `toml:"configs,omitempty"`
}

type containerdConfigPluginsCRIRegistryConfig struct {
	Auth containerdConfigPluginsCRIRegistryConfigAuth `toml:"auth,omitempty"`
}

type containerdConfigPluginsCRIRegistryConfigAuth struct {
	Username string `toml:"username,omitempty"`
	Password string `toml:"password,omitempty"`
	Token    string `toml:"token,omitempty"`
}

type containerdHostsConfig struct {
	Server string                               `toml:"server,omitempty"`
	Host   map[string]containerdHostsConfigHost `toml:"hosts,omitempty"`
}

type containerdHostsConfigHost struct {
	Capabilities []string `toml:"capabilities,omitempty"`
	SkipVerify   bool     `toml:"skip_verify,omitempty"`
	OverridePath bool     `toml:"override_path,omitempty"`
}

func containerdAuthConfig(registries []types.ContainerdRegistry) containerdConfig {
	authConfigs := make(map[string]containerdConfigPluginsCRIRegistryConfig, len(registries))
	for _, registry := range registries {
		if registry.Username != "" || registry.Password != "" || registry.Token != "" {
			for _, url := range registry.URLs {
				authConfigs[url] = containerdConfigPluginsCRIRegistryConfig{
					Auth: containerdConfigPluginsCRIRegistryConfigAuth{
						Username: registry.Username,
						Password: registry.Password,
						Token:    registry.Token,
					},
				}
			}
		}
	}

	return containerdConfig{
		Version: 2,
		Plugins: containerdConfigPlugins{
			CRI: containerdConfigPluginsCRI{
				Registry: containerdConfigPluginsCRIRegistry{
					Configs: authConfigs,
				},
			},
		},
	}
}

func containerdHostConfig(registry types.ContainerdRegistry) containerdHostsConfig {
	if len(registry.URLs) == 0 {
		return containerdHostsConfig{}
	}

	hosts := make(map[string]containerdHostsConfigHost, len(registry.URLs))
	for _, url := range registry.URLs {
		hosts[url] = containerdHostsConfigHost{
			Capabilities: []string{"pull", "resolve"},
			SkipVerify:   registry.SkipVerify,
			OverridePath: registry.OverridePath,
		}
	}

	return containerdHostsConfig{
		Server: registry.URLs[0],
		Host:   hosts,
	}
}

// Containerd configures configuration and arguments for containerd on the local node.
// Optionally, a number of registry mirrors and auths can be configured.
func Containerd(snap snap.Snap, registries []types.ContainerdRegistry) error {
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
		PauseImage:        "ghcr.io/canonical/k8s-snap/pause:3.10",
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

	// registry auths
	if authConfig := containerdAuthConfig(registries); len(authConfig.Plugins.CRI.Registry.Configs) > 0 {
		b, err := toml.Marshal(authConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal registry auth configurations: %w", err)
		}

		if err := os.WriteFile(path.Join(snap.ContainerdExtraConfigDir(), "k8sd-auths.toml"), b, 0600); err != nil {
			return fmt.Errorf("failed to write registry auth configurations: %w", err)
		}
	}

	// registry mirrors
	for _, registry := range registries {
		if hostConfig := containerdHostConfig(registry); len(hostConfig.Host) > 0 {
			b, err := toml.Marshal(hostConfig)
			if err != nil {
				return fmt.Errorf("failed to render registry mirrors for %s: %w", registry.Host, err)
			}

			dir := path.Join(snap.ContainerdRegistryConfigDir(), registry.Host)
			if err := os.Mkdir(dir, 0700); err != nil && !os.IsExist(err) {
				return fmt.Errorf("failed to create directory for registry %s: %w", registry.Host, err)
			}
			if err := os.WriteFile(path.Join(dir, "hosts.toml"), b, 0600); err != nil {
				return fmt.Errorf("failed to write hosts.toml for registry %s: %w", registry.Host, err)
			}
		}
	}

	return nil
}
