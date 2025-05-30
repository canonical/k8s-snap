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
	log := log.FromContext(ctx)

	for _, pluginDir := range PluginDirs {
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
