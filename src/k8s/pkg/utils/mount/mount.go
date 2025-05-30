package mountutils

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/canonical/k8s/pkg/log"
)

type MountManager interface {
	ForEachMount(ctx context.Context, callback func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error) error
	Unmount(ctx context.Context, mountPoint string, flags int) error
	DetachLoopDevice(ctx context.Context, loopDevice string) error
}

func forEachMount(ctx context.Context, mountsPath string, callback func(ctx context.Context, device string, mountPoint string, fsType string, flags string) error) error {
	log := log.FromContext(ctx)

	file, err := os.Open(mountsPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", mountsPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) < 4 {
			log.Error(fmt.Errorf("less than 4 fields in mount entry: %s", line), "skipping invalid mount entry")
			continue
		}

		if err := callback(ctx, fields[0], fields[1], fields[2], fields[3]); err != nil {
			log.Error(err, "callback failed for mount entry", "device", fields[0], "mountPoint", fields[1], "fsType", fields[2], "flags", fields[3])
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read %s: %w", mountsPath, err)
	}

	return nil
}
