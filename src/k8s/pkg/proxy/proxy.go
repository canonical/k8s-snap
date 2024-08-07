package proxy

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/canonical/k8s/pkg/log"
)

func startProxy(ctx context.Context, listenURL string, endpointURLs []string) error {
	if len(endpointURLs) == 0 {
		return fmt.Errorf("empty list of endpoints")
	}
	srvs := make([]*net.SRV, len(endpointURLs))
	for i, endpoint := range endpointURLs {
		if u, err := url.Parse(endpoint); err == nil && u.Host != "" {
			endpoint = u.Host
		}
		host, port, err := net.SplitHostPort(endpoint)
		if err != nil {
			return fmt.Errorf("failed to parse endpoint %q: %w", endpoint, err)
		}
		portNumber, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return fmt.Errorf("failed to parse port %q: %w", port, err)
		}
		srvs[i] = &net.SRV{Target: host, Port: uint16(portNumber)}
	}

	l, err := net.Listen("tcp", listenURL)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	p := tcpproxy{
		Listener:        l,
		Endpoints:       srvs,
		MonitorInterval: time.Minute,
	}

	log := log.FromContext(ctx).WithValues(
		"controller", "proxy",
		"address", listenURL,
		"endpoints", endpointURLs,
	)
	log.Info("Starting proxy")
	go func() {
		if err := p.Run(); err != nil {
			log.Error(err, "Proxy failed")
		}
	}()

	<-ctx.Done()
	p.Stop()

	return nil
}
