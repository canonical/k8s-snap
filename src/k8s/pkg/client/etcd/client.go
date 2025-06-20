package etcd

import (
	"context"
	"fmt"
	"path/filepath"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

type externalClient struct {
	storageDir string
	endpoints  []string
	etcdClient *clientv3.Client
}

func NewExternalClient(storageDir string, endpoints []string) (*externalClient, error) {
	etcdClient, err := newClient(storageDir, endpoints)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &externalClient{
		storageDir: storageDir,
		endpoints:  endpoints,
		etcdClient: etcdClient,
	}, nil
}

func newClient(dir string, endpoints []string) (*clientv3.Client, error) {
	config, err := embed.ConfigFromFile(filepath.Join(dir, "etcd.yaml"))
	if err != nil {
		return nil, fmt.Errorf("failed to load etcd config: %w", err)
	}

	tlsConfig, err := config.ClientTLSInfo.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client TLS config: %w", err)
	}

	return clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		TLS:       tlsConfig,
	})
}

func (c *externalClient) RemoveNodeByAddress(ctx context.Context, peerURL string) error {
	// List all members
	resp, err := c.etcdClient.MemberList(ctx)
	if err != nil {
		return fmt.Errorf("failed to list etcd members: %w", err)
	}

	// Find the member with the matching peerURL
	var memberID uint64
	for _, m := range resp.Members {
		for _, url := range m.PeerURLs {
			if url == peerURL {
				memberID = m.ID
				break
			}
		}
	}

	if memberID == 0 {
		return fmt.Errorf("no etcd member found with peer URL: %s", peerURL)
	}

	// Remove the member
	_, err = c.etcdClient.MemberRemove(ctx, memberID)
	if err != nil {
		return fmt.Errorf("failed to remove etcd member: %w", err)
	}

	return nil
}
