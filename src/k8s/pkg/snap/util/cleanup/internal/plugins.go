package internal

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/canonical/k8s/pkg/log"
)

var PluginDirs = []string{"/var/lib/kubelet/plugins/", "/var/lib/kubelet/plugins_registry/"}

func RemovePluginSockets(ctx context.Context) {
	for _, pluginDir := range PluginDirs {
		removeSockets(ctx, pluginDir)
	}
}

func removeSockets(ctx context.Context, dir string) {
	log := log.FromContext(ctx)

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Error(err, "failed to list files under directory", "dir", dir)
		return
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			removeSockets(ctx, path)
			continue
		}

		if strings.HasSuffix(entry.Name(), ".sock") {
			if err := os.RemoveAll(path); err != nil {
				log.Error(err, "failed to remove socket", "path", path)
				continue
			}
		}
	}
}
