package dqlite

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/canonical/go-dqlite/app"
	"github.com/canonical/go-dqlite/client"
)

type ClientOpts struct {
	// ClusterYAML is the path cluster.yaml, containing the list of known dqlite nodes.
	ClusterYAML string
	// ClusterCert is the path to cluster.crt, containing the dqlite cluster certificate.
	ClusterCert string
	// ClusterKey is the path to cluster.key, containing the dqlite cluster private key.
	ClusterKey string
}

type Client struct {
	// clientGetter dynamically creates a dqlite client. This is because the dqlite client
	// must dynamically connect to the leader node of the cluster.
	clientGetter func(context.Context) (*client.Client, error)
}

// NewClient creates a new client connected to the leader of the dqlite cluster.
func NewClient(ctx context.Context, opts ClientOpts) (*Client, error) {
	var options []client.Option
	if opts.ClusterCert != "" && opts.ClusterKey != "" {
		cert, err := tls.LoadX509KeyPair(opts.ClusterCert, opts.ClusterKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load x509 keypair from certificate %q and key %q: %w", opts.ClusterCert, opts.ClusterKey, err)
		}
		b, err := os.ReadFile(opts.ClusterCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read x509 certificate: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("bad certificate in %q", opts.ClusterCert)
		}
		options = append(options, client.WithDialFunc(client.DialFuncWithTLS(client.DefaultDialFunc, app.SimpleDialTLSConfig(cert, pool))))
	}

	return &Client{
		clientGetter: func(ctx context.Context) (*client.Client, error) {
			store, err := client.NewYamlNodeStore(opts.ClusterYAML)
			if err != nil {
				return nil, fmt.Errorf("failed to open node store from %q: %w", opts.ClusterYAML, err)
			}
			c, err := client.FindLeader(ctx, store, options...)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to dqlite leader: %w", err)
			}
			return c, nil
		},
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	client, err := c.clientGetter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create dqlite client: %w", err)
	}
	if err = client.Close(); err != nil {
		return fmt.Errorf("failed to close dqlite client: %w", err)
	}
	return nil
}
