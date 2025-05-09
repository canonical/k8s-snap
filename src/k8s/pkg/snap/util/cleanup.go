package snaputil

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils/shims"
	"golang.org/x/sys/unix"
)

// TryCleanupContainerdPaths attempts to clean up all containerd directories which were
// created by the k8s-snap based on the existence of their respective lockfiles
// located in the directory returned by `s.LockFilesDir()`.
func TryCleanupContainerdPaths(ctx context.Context, s snap.Snap) {
	log := log.FromContext(ctx)
	for lockpath, dirpath := range ContainerdLockPathsForSnap(s) {
		// Ensure lockfile exists:
		log.Info("Cleaning up containerd data directory", "directory", dirpath)
		if _, err := os.Stat(lockpath); os.IsNotExist(err) {
			log.Info("WARN: failed to find containerd lockfile, no cleanup will be perfomed", "lockfile", lockpath, "directory", dirpath)
			continue
		}

		// Ensure lockfile's contents is the one we expect:
		lockfile_contents := ""
		if contents, err := os.ReadFile(lockpath); err != nil {
			log.Info("WARN: failed to read contents of lockfile", "lockfile", lockpath, "error", err)
			continue
		} else {
			lockfile_contents = string(contents)
		}

		if lockfile_contents != dirpath {
			log.Info("WARN: lockfile points to different path than expected", "lockfile", lockpath, "expected", dirpath, "actual", lockfile_contents)
			continue
		}

		// Check directory exists before attempting to remove:
		if stat, err := os.Stat(dirpath); os.IsNotExist(err) {
			log.Info("Containerd directory doesn't exist; skipping cleanup", "directory", dirpath)
		} else {
			realPath := dirpath
			if stat.Mode()&fs.ModeSymlink != 0 {
				// NOTE(aznashwan): because of the convoluted interfaces-based way the snap
				// composes and creates the original lockfiles (see k8sd/setup/containerd.go)
				// this check is meant to defend against accidental code/configuration errors which
				// might lead to the root FS being deleted:
				realPath, err = os.Readlink(dirpath)
				if err != nil {
					log.Error(err, fmt.Sprintf("Failed to os.Readlink the directory path for lockfile %q pointing to %q. Skipping cleanup", lockpath, dirpath))
					continue
				}
			}

			if realPath == "/" {
				log.Error(fmt.Errorf("There is some configuration/logic error in the current versions of the k8s-snap related to lockfile %q (meant to lock %q, which points to %q) which could lead to accidental wiping of the root file system.", lockpath, dirpath, realPath), "Please report this issue upstream immediately.")
				continue
			}

			if err := os.RemoveAll(dirpath); err != nil {
				log.Info("WARN: failed to remove containerd data directory", "directory", dirpath, "error", err, "realPath", realPath)
				continue // Avoid removing the lockfile path.
			}
		}

		if err := os.Remove(lockpath); err != nil {
			log.Info("WARN: Failed to remove containerd lockfile", "lockfile", lockpath)
		}
	}
}

// [DANGER] Cleanup containers and runtime state. Note that the order of operations below is crucial.
// Cleanup is done on a best-effort basis, and errors are logged but not returned.
func TryCleanupContainers(ctx context.Context, s snap.Snap) {
	removeContainers(ctx)
	removeNetworkNamespaces(ctx)
	removeVolumeMountsGracefully(ctx)
	removeVolumeMountsForce(ctx)
	removePluginSockets(ctx)
	removeLoopDevices(ctx)
}

func removeContainers(ctx context.Context) {
	log := log.FromContext(ctx)
	pids, err := shims.RunningContainerdShimPIDs(ctx)
	if err != nil {
		log.Error(err, "failed to get containerd shim PIDs")
		return
	}

	for _, pid := range pids {
		intPid, err := strconv.Atoi(pid)
		if err != nil {
			log.Error(err, "failed to convert PID to integer", "pid", pid)
			continue
		}

		process, err := os.FindProcess(intPid)
		if err != nil {
			log.Error(err, "failed to find containerd shim PID", "pid", intPid)
			continue
		}

		if err := process.Kill(); err != nil {
			log.Error(err, "failed to kill containerd shim PID", "pid", intPid)
			continue
		}
	}
}

func removeNetworkNamespaces(ctx context.Context) {
	log := log.FromContext(ctx)

	netnsDir := "/run/netns"
	entries, err := os.ReadDir(netnsDir)
	if err != nil {
		log.Error(err, "failed to list files under network namespace directory", "netnsDir", netnsDir)
		return
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cni-") {
			nsPath := filepath.Join(netnsDir, entry.Name())

			if err := unix.Unmount(nsPath, unix.MNT_DETACH); err != nil {
				log.Error(err, "failed to unmount network namespace", "namespace", entry.Name())
				continue
			}

			if err := os.Remove(nsPath); err != nil {
				log.Error(err, "failed to remove network namespace", "namespace", entry.Name())
				continue
			}
		}
	}
}

func removeLoopDevices(ctx context.Context) {
	log := log.FromContext(ctx)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Error(err, "failed to open /proc/mounts")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		device := fields[0]
		mountPoint := fields[1]

		if strings.HasPrefix(mountPoint, "/var/lib/kubelet/pods") {
			// Gather loop devices for detachment
			if strings.HasPrefix(device, "/dev/loop") {
				cmd := exec.CommandContext(ctx, "losetup", "-d", device)
				if err := cmd.Run(); err != nil {
					log.Error(err, "failed to detach loop device using losetup", "device", device)
					continue
				}

			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Error(err, "failed to read /proc/mounts")
	}
}

func removeVolumeMountsGracefully(ctx context.Context) {
	log := log.FromContext(ctx)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Error(err, "failed to open /proc/mounts")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		mountPoint := fields[1]
		mountType := fields[2]

		for _, prefix := range []string{"/var/lib/kubelet/pods", "/run/containerd/io.containerd.", "/var/snap/k8s/common/var/lib/containerd/"} {
			if strings.HasPrefix(mountPoint, prefix) {
				// unmount Pod NFS volumes only forcefully, as unmounting them normally may hang otherwise.
				if !strings.Contains(mountType, "nfs") {
					// unmount Pod volumes gracefully.
					if err := unix.Unmount(mountPoint, unix.MNT_DETACH); err != nil {
						log.Error(err, "failed to unmount mount point", "mountPoint", mountPoint)
						continue
					}
				}
			}
		}

	}

	if err := scanner.Err(); err != nil {
		log.Error(err, "failed to read /proc/mounts")
	}
}

func removeVolumeMountsForce(ctx context.Context) {
	log := log.FromContext(ctx)

	file, err := os.Open("/proc/mounts")
	if err != nil {
		log.Error(err, "failed to open /proc/mounts")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		mountPoint := fields[1]

		for _, prefix := range []string{"/var/lib/kubelet/pods", "/run/containerd/io.containerd.", "/var/lib/kubelet/plugins", "/var/snap/k8s/common/var/lib/containerd/"} {
			if strings.HasPrefix(mountPoint, prefix) {
				// unmount lingering Pod volumes by force, to prevent potential volume leaks.
				if err := unix.Unmount(mountPoint, unix.MNT_FORCE); err != nil {
					log.Error(err, "failed to force unmount mount point", "mountPoint", mountPoint)
					continue
				}
			}
		}

	}

	if err := scanner.Err(); err != nil {
		log.Error(err, "failed to read /proc/mounts")
	}
}

func removePluginSockets(ctx context.Context) {
	log := log.FromContext(ctx)

	for _, pluginDir := range []string{"/var/lib/kubelet/plugins/", "/var/lib/kubelet/plugins_registry/"} {
		entries, err := os.ReadDir(pluginDir)
		if err != nil {
			log.Error(err, "failed to list files under plugin directory", "pluginDir", pluginDir)
			continue
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".sock") {
				socketsPath := filepath.Join(pluginDir, entry.Name())
				if err := os.RemoveAll(socketsPath); err != nil {
					log.Error(err, "failed to remove socket", "socketPath", socketsPath)
					continue
				}
			}
		}
	}
}

func RemoveKubeProxyRules(ctx context.Context, s snap.Snap) {
	log := log.FromContext(ctx)

	// Remove kube-proxy rules
	cmd := exec.CommandContext(ctx, filepath.Join(s.K8sBinDir(), "kube-proxy"), "--cleanup")

	if err := cmd.Run(); err != nil {
		log.Error(err, "failed to run kube-proxy cleanup")
	}
}
