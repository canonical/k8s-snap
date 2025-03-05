package setup

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/canonical/k8s/pkg/k8sd/images"
	"github.com/canonical/k8s/pkg/snap"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/pelletier/go-toml"
)

const defaultPauseImage = "ghcr.io/canonical/k8s-snap/pause:3.10"

func defaultContainerdConfig(
	cniConfDir string,
	cniBinDir string,
	importsDir string,
	registryConfigDir string,
	pauseImage string,
) map[string]any {
	return map[string]any{
		"version":   2,
		"oom_score": 0,
		"imports":   []string{filepath.Join(importsDir, "*.toml")},

		"grpc": map[string]any{
			"uid":                   0,
			"gid":                   0,
			"max_recv_message_size": 16777216,
			"max_send_message_size": 16777216,
		},

		"debug": map[string]any{
			"uid":     0,
			"gid":     0,
			"address": "",
			"level":   "",
		},

		"metrics": map[string]any{
			"address":        "",
			"grpc_histogram": false,
		},

		"cgroup": map[string]any{
			"path": "",
		},

		"plugins": map[string]any{
			"io.containerd.grpc.v1.cri": map[string]any{
				"stream_server_address":       "127.0.0.1",
				"stream_server_port":          "0",
				"enable_selinux":              false,
				"sandbox_image":               pauseImage,
				"stats_collect_period":        10,
				"enable_tls_streaming":        false,
				"max_container_log_line_size": 16384,

				"containerd": map[string]any{
					"no_pivot":             false,
					"default_runtime_name": "runc",

					"runtimes": map[string]any{
						"runc": map[string]any{
							"runtime_type": "io.containerd.runc.v2",
							"options": map[string]any{
								"SystemdCgroup": true,
							},
						},
					},
				},

				"cni": map[string]any{
					"bin_dir":  cniBinDir,
					"conf_dir": cniConfDir,
				},

				"registry": map[string]any{
					"config_path": registryConfigDir,
				},
			},
		},
	}
}

// Containerd configures configuration and arguments for containerd on the local node.
// Optionally, a number of registry mirrors and auths can be configured.
func Containerd(snap snap.Snap, extraContainerdConfig map[string]any, extraArgs map[string]*string) error {
	// We create the directories here since PreInitCheck is called before this
	// This ensures we only create the directories if we are going to configure containerd
	for _, dir := range []string{
		snap.ContainerdConfigDir(),
		snap.ContainerdExtraConfigDir(),
		snap.ContainerdRegistryConfigDir(),
	} {
		if dir == "" {
			continue
		}
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("failed to create required directory: %w", err)
		}
	}

	configToml := defaultContainerdConfig(
		snap.CNIConfDir(),
		snap.CNIBinDir(),
		snap.ContainerdExtraConfigDir(),
		snap.ContainerdRegistryConfigDir(),
		defaultPauseImage,
	)

	if err := mergo.Merge(&configToml, extraContainerdConfig, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return fmt.Errorf("failed to merge containerd config.toml overrides: %w", err)
	}

	b, err := toml.Marshal(configToml)
	if err != nil {
		return fmt.Errorf("failed to render containerd config.toml: %w", err)
	}

	if err := utils.WriteFile(filepath.Join(snap.ContainerdConfigDir(), "config.toml"), b, 0o600); err != nil {
		return fmt.Errorf("failed to write config.toml: %w", err)
	}

	if _, err := snaputil.UpdateServiceArguments(snap, "containerd", map[string]string{
		"--address": snap.ContainerdSocketPath(),
		"--config":  filepath.Join(snap.ContainerdConfigDir(), "config.toml"),
		"--root":    snap.ContainerdRootDir(),
		"--state":   snap.ContainerdStateDir(),
	}, nil); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	// Apply extra arguments after the defaults, so they can override them.
	updateArgs, deleteArgs := utils.ServiceArgsFromMap(extraArgs)
	if _, err := snaputil.UpdateServiceArguments(snap, "containerd", updateArgs, deleteArgs); err != nil {
		return fmt.Errorf("failed to write arguments file: %w", err)
	}

	cniBinary := filepath.Join(snap.CNIBinDir(), "cni")
	if err := utils.CopyFile(snap.CNIPluginsBinary(), cniBinary); err != nil {
		return fmt.Errorf("failed to copy cni plugin binary: %w", err)
	}
	if err := os.Chmod(cniBinary, 0o700); err != nil {
		return fmt.Errorf("failed to chmod cni plugin binary: %w", err)
	}
	if err := os.Chown(cniBinary, snap.UID(), snap.GID()); err != nil {
		return fmt.Errorf("failed to chown cni plugin binary: %w", err)
	}

	// for each of the CNI plugins, ensure they are a symlink to the "cni" binary
	for _, plugin := range snap.CNIPlugins() {
		pluginInstallPath := filepath.Join(snap.CNIBinDir(), plugin)

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

	if err := saveSnapContainerdPaths(snap); err != nil {
		return err
	}

	return nil
}

// ContainerdLockPathsForSnap returns a mapping between the absolute paths of
// the lockfiles within the k8s snap and the absolute paths of the containerd
// directory they lock.
//
// WARN: these lockfiles are meant to be used in later cleanup stages.
// DO NOT include any system paths which are not managed by the k8s-snap!
//
// It intentionally does NOT include the containerd base dir lockfile
// (which most of the rest of the paths are based on), as it is meant
// to indicate the root of the containerd install ('/' or '/var/snap/k8s/*').
func ContainerdLockPathsForSnap(s snap.Snap) map[string]string {
	m := map[string]string{
		"containerd-socket-path": s.ContainerdSocketDir(),
		"containerd-config-dir":  s.ContainerdConfigDir(),
		"containerd-root-dir":    s.ContainerdRootDir(),
		"containerd-cni-bin-dir": s.CNIBinDir(),
	}

	prefixed := map[string]string{}
	for k, v := range m {
		prefixed[filepath.Join(s.LockFilesDir(), k)] = v
	}

	return prefixed
}

// saveSnapContainerdPaths creates the lock files for the containerd directory paths to be used for later cleanup.
func saveSnapContainerdPaths(s snap.Snap) error {
	for lockpath, dirpath := range ContainerdLockPathsForSnap(s) {
		if err := utils.WriteFile(lockpath, []byte(dirpath), 0o600); err != nil {
			return fmt.Errorf("failed to write %s: %w", lockpath, err)
		}
	}

	// Save the Containerd Base Dir separately:
	baseDirPath := filepath.Join(s.LockFilesDir(), snap.ContainerdBaseDir)
	if err := utils.WriteFile(baseDirPath, []byte(s.GetContainerdBaseDir()), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", baseDirPath, err)
	}

	return nil
}

func init() {
	images.Register(defaultPauseImage)
}
