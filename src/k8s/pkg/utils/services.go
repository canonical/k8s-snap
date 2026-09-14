package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ServiceArgsFromMap processes a map of string pointers and categorizes them into update and delete lists.
// - If the value pointer is nil, it adds the argument name to the delete list.
// - If the value pointer is not nil, it adds the argument and its value to the update map.
func ServiceArgsFromMap(args map[string]*string) (map[string]string, []string) {
	updateArgs := make(map[string]string)
	deleteArgs := make([]string, 0)

	for arg, val := range args {
		if val == nil {
			deleteArgs = append(deleteArgs, arg)
		} else {
			updateArgs[arg] = *val
		}
	}
	return updateArgs, deleteArgs
}

var ErrUnitNotRunning = errors.New("unit is not running")

// RunningServiceArgs queries systemd for the MainPID of the snap service unit
// and returns its parsed command-line arguments from /proc/<pid>/cmdline.
// Returns nil, nil when the service is not running.
func RunningServiceArgs(ctx context.Context, serviceName string) (map[string]string, error) {
	unitName := fmt.Sprintf("snap.k8s.%s.service", serviceName)

	// unit not found:
	// LoadState=not-found
	// ActiveState=inactive

	// unit not running
	// LoadState=loaded
	// ActiveState=inactive

	// unit running
	// LoadState=loaded
	// ActiveState=active

	loadState, err := getUnitLoadState(ctx, unitName)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit %q LoadState: %w", unitName, err)
	}

	if loadState == LoadStateNotFound {
		return nil, fmt.Errorf("unit %q was not found", unitName)
	}

	// unit is loaded, let's see if it's running

	activeState, err := getUnitActiveState(ctx, unitName)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit %q ActiveState: %w", unitName, err)
	}

	if activeState == ActiveStateInactive {
		return nil, fmt.Errorf("%q: %w", unitName, ErrUnitNotRunning)
	}

	// unit is loaded and running, let's get its pid

	pid, err := getUnitMainPID(ctx, unitName)
	if err != nil {
		return nil, fmt.Errorf("failed to get unit %q MainPID: %w", unitName, err)
	}

	if pid == 0 {
		// this should only happen in case of a race condition.
		// for snap services with Type=simple or Type=notify (containerd),
		//  MainPID is only 0 if the unit is either inactive or not-found.
		// since we're making sure the service is loaded and active before getting the PID,
		// the only possible explanation for this branch is that the service was stopped right after
		// we checked the load and active states.
		return nil, fmt.Errorf("pid is 0 for unit %q: %w", unitName, ErrUnitNotRunning)
	}

	cmdlineData, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// this branch is the similar to pid == 0. it shouldn't really happen, unless due to a race condition.
			return nil, fmt.Errorf("cmdline file doesn't exist for unit %q (pid %d): %w", unitName, pid, ErrUnitNotRunning)
		}
		return nil, fmt.Errorf("failed to read cmdline for pid %d (unit %q): %w", pid, unitName, err)
	}

	args := make(map[string]string)
	for _, part := range bytes.Split(cmdlineData, []byte{0})[1:] { // skip argv[0] (binary path)
		if len(part) == 0 {
			continue
		}
		key, value := ParseArgumentLine(string(part))
		if key != "" {
			args[key] = value
		}
	}
	return args, nil
}

type LoadState string

const (
	LoadStateLoaded   = "loaded"
	LoadStateNotFound = "not-found"
)

func getUnitLoadState(ctx context.Context, unit string) (LoadState, error) {
	out, err := exec.CommandContext(ctx, "systemctl", "show", unit, "--property=LoadState", "--value").Output()
	if err != nil {
		return "", fmt.Errorf("failed to query LoadState for unit %q: %w", unit, err)
	}

	switch st := strings.TrimSpace(string(out)); st {
	case "loaded":
		return LoadStateLoaded, nil
	case "not-found":
		return LoadStateNotFound, nil
	default:
		return "", fmt.Errorf("invalid LoadState %q for unit %q", st, unit)
	}
}

type ActiveState string

const (
	ActiveStateActive   ActiveState = "active"
	ActiveStateInactive ActiveState = "inactive"
)

func getUnitActiveState(ctx context.Context, unit string) (ActiveState, error) {
	out, err := exec.CommandContext(ctx, "systemctl", "show", unit, "--property=ActiveState", "--value").Output()
	if err != nil {
		return "", fmt.Errorf("failed to query ActiveState for unit %q: %w", unit, err)
	}

	switch st := strings.TrimSpace(string(out)); st {
	case "active":
		return ActiveStateActive, nil
	case "inactive":
		return ActiveStateInactive, nil
	default:
		return "", fmt.Errorf("invalid ActiveState %q for unit %q", st, unit)
	}
}

func getUnitMainPID(ctx context.Context, unit string) (int, error) {
	out, err := exec.CommandContext(ctx, "systemctl", "show", unit, "--property=MainPID", "--value").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to query MainPID for unit %q: %w", unit, err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("failed to convert unit %q pid %q to int: %w", unit, string(out), err)
	}

	return pid, nil
}
