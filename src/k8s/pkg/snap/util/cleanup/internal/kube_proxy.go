package internal

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
)

func RemoveKubeProxyRules(ctx context.Context, s snap.Snap) {
	log := log.FromContext(ctx)

	// Remove kube-proxy rules
	cmd := exec.CommandContext(ctx, filepath.Join(s.K8sBinDir(), "kube-proxy"), "--cleanup")

	if err := cmd.Run(); err != nil {
		log.Error(err, "failed to run kube-proxy cleanup")
	}
}
