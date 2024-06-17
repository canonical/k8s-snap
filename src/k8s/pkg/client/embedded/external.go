package embedded

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// externalClient implements Client using `k8s-dqlite embeddedctl` commands.
type externalClient struct {
	binary     string
	storageDir string
}

func NewExternalClient(binary string, storageDir string) *externalClient {
	return &externalClient{binary: binary, storageDir: storageDir}
}

func (c *externalClient) RemoveNodeByAddress(ctx context.Context, peerURL string) error {
	command := []string{c.binary, "embeddedctl", "member", "remove", "--storage-dir", c.storageDir, "--peer-url", peerURL}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	if b, err := cmd.CombinedOutput(); err != nil && !bytes.Contains(b, []byte("cluster member not found")) {
		return fmt.Errorf("command failed, rc=%v command=%v output=%q", cmd.ProcessState.ExitCode(), command, string(b))
	}
	return nil
}
