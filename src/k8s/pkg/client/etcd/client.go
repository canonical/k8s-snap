package etcd

import (
	"context"
	"fmt"

	pkiutil "github.com/canonical/k8s/pkg/utils/pki"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	*clientv3.Client
}

func NewClient(pkiDir string, endpoints []string) (*Client, error) {
	certFile := fmt.Sprintf("%s/server.crt", pkiDir)
	keyFile := fmt.Sprintf("%s/server.key", pkiDir)
	caFile := fmt.Sprintf("%s/ca.crt", pkiDir)

	tlsConfig, err := pkiutil.LoadTLSConfigFromPath(certFile, keyFile, caFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		TLS:       tlsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Client{
		client,
	}, nil
}

func (c *Client) RemoveNodeByAddress(ctx context.Context, peerURL string) error {
	resp, err := c.MemberList(ctx)
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

	if _, err = c.MemberRemove(ctx, memberID); err != nil {
		return fmt.Errorf("failed to remove etcd member: %w", err)
	}

	return nil
}
