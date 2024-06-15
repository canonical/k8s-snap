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
	cmd := exec.CommandContext(ctx, c.binary, "embeddedctl", "member", "remove", "--storage-dir", c.storageDir, "--peer-url", peerURL)
	if b, err := cmd.CombinedOutput(); err != nil && !bytes.Contains(b, []byte("cluster member not found")) {
		return fmt.Errorf("command failed, rc=%v output=%q", cmd.ProcessState.ExitCode(), string(b))
	}
	return nil
}
