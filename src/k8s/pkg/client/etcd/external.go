package etcd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// externalClient implements Client using `k8s-dqlite dbctl` commands.
type externalClient struct {
	binary     string
	storageDir string
}

func NewExternalClient(binary string, storageDir string) *externalClient {
	return &externalClient{binary: binary, storageDir: storageDir}
}

func (c *externalClient) RemoveNodeByAddress(ctx context.Context, peerURL string) error {
	command := []string{c.binary, "dbctl", "member", "remove", "--storage-dir", c.storageDir, "--peer-url", peerURL}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	b, err := cmd.CombinedOutput()
	switch {
	case err == nil:
		// command succeeded
		return nil
	case bytes.Contains(b, []byte("cluster member not found")):
		// member does not exist, no error
		return nil
	case bytes.Contains(b, []byte("etcdserver: server stopped")):
		// member remove will sometimes fail while removing itself
		return nil
	default:
		return fmt.Errorf("command failed, rc=%v command=%v output=%q", cmd.ProcessState.ExitCode(), command, string(b))
	}
}
