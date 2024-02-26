package shims

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// reBinaryName is the regular expression used to match containerd shim and /pause processes.
var reBinaryName = regexp.MustCompile(`(^/snap/k8s/.*/bin/containerd-shim-runc-v2|^/pause$)`)

// RunningContainerdShimPIDs returns a list of all the pids on the system that have been started by a containerd shim.
func RunningContainerdShimPIDs(ctx context.Context) ([]string, error) {
	procs, err := listAllSystemProcesses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list running host processes: %w", err)
	}

	return findAllChildren(findShimPIDs(procs), makeShallowChildPIDs(procs)), nil
}

type processInfo struct {
	command   string
	parentPID string
}

// listAllSystemProcesses returns a map of all running processes on the host.
// for each process, we store the ppid and the command line.
func listAllSystemProcesses(ctx context.Context) (map[string]processInfo, error) {
	// output is a list of lines in the following format, one line for each running process:
	// [pid] [ppid] [arg1 arg2 arg3 ...]
	// [pid] [ppid] [arg1 arg2 arg3 ...]
	stdout, err := exec.CommandContext(ctx, "bash", "-c", `ps -e -o pid=,ppid=,args=`).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute ps command (output=%q): %w", stdout, err)
	}

	result := map[string]processInfo{}
	for _, line := range strings.Split(string(stdout), "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		result[parts[0]] = processInfo{
			parentPID: parts[1],
			command:   strings.Join(parts[2:], " "),
		}
	}

	return result, nil
}

// makeShallowChildPIDs returns a shallow map of the direct children for each process.
func makeShallowChildPIDs(procs map[string]processInfo) map[string][]string {
	result := make(map[string][]string, len(procs))
	for pid, info := range procs {
		result[info.parentPID] = append(result[info.parentPID], pid)
	}
	return result
}

// findShimPIDs returns the list of PIDs of the parent shim and pause processes.
func findShimPIDs(procs map[string]processInfo) []string {
	var result []string
	for pid, info := range procs {
		if reBinaryName.MatchString(info.command) {
			result = append(result, pid)
		}
	}
	return result
}

// findAllChildren returns a list of all process IDs starting from a given set of parents.
func findAllChildren(startPIDs []string, shallowChildPIDs map[string][]string) []string {
	var result []string
	for _, pid := range startPIDs {
		result = append(result, pid)
		result = append(result, findAllChildren(shallowChildPIDs[pid], shallowChildPIDs)...)
	}
	return result
}
