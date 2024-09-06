package snapd

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

var socketPath = "/run/snapd.socket"

type Client struct {
	client *http.Client
}

func NewClient() (*Client, error) {
	defaultTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("http.DefaultTransport is not a *http.Transport")
	}

	unixTransport := defaultTransport.Clone()
	defaultDialContext := unixTransport.DialContext

	unixTransport.DialContext = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return defaultDialContext(ctx, "unix", socketPath)
	}

	return &Client{client: &http.Client{Transport: unixTransport}}, nil
}
