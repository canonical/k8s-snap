package component

import (
	"context"
	"fmt"
	"strings"
	"time"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/utils"
)

func EnableDNSComponent(request api.UpdateDNSComponentRequest) error {
	manager, err := NewManager()
	if err != nil {
		return fmt.Errorf("failed to get component manager: %w", err)
	}

	var serviceIP string
	upstreamNameserver := "/etc/resolv.conf"
	clusterDomain := "cluster.local"
	if request.Config != nil {
		config := request.Config
		if len(config.UpstreamNameservers) > 0 {
			upstreamNameserver = strings.Join(config.UpstreamNameservers, " ")
		}

		if config.ClusterDomain != "" {
			clusterDomain = config.ClusterDomain
		}

		serviceIP = config.ServiceIP
	}

	values := map[string]any{
		"servers": []map[string]any{
			{
				"zones": []map[string]any{
					{"zone": "."},
				},
				"port": 53,
				"plugins": []map[string]any{
					{"name": "errors"},
					{"name": "health", "configBlock": "lameduck 5s"},
					{"name": "ready"},
					{
						"name":        "kubernetes",
						"parameters":  fmt.Sprintf("%s in-addr.arpa ip6.arpa", clusterDomain),
						"configBlock": "pods insecure\nfallthrough in-addr.arpa ip6.arpa\nttl 30",
					},
					{"name": "prometheus", "parameters": "0.0.0.0:9153"},
					{"name": "forward", "parameters": fmt.Sprintf(". %s", upstreamNameserver)},
					{"name": "cache", "parameters": "30"},
					{"name": "loop"},
					{"name": "reload"},
					{"name": "loadbalance"},
				},
			},
		},
	}

	if serviceIP != "" {
		values["service"] = map[string]any{
			"clusterIP": serviceIP,
		}
	}

	err = manager.EnableWithValues("dns", values)
	if err != nil {
		return fmt.Errorf("failed to enable dns component: %w", err)
	}

	client, err := utils.NewKubeClient("/etc/kubernetes/admin.conf")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	svc, err := client.GetService(ctx, "ck-dns-coredns", "kube-system")
	if err != nil {
		return fmt.Errorf("failed to get dns service: %w", err)
	}

	dnsIP := svc.Spec.ClusterIP

	err = utils.UpdateServiceArgs("cluster-dns", dnsIP, "kubelet")
	if err != nil {
		return fmt.Errorf("failed to update cluster-dns argument: %w", err)
	}

	err = utils.UpdateServiceArgs("cluster-domain", clusterDomain, "kubelet")
	if err != nil {
		return fmt.Errorf("failed to update cluster-domain argument: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = utils.RestartService(ctx, "kubelet")
	if err != nil {
		return fmt.Errorf("failed to restart service 'kubelet': %w", err)
	}

	return nil
}
