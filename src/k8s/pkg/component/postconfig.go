package component

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/k8s/pkg/utils"
)

func ExecuteDNSPostConfig(values map[string]any) error {
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

	clusterDomain, err := utils.GetServiceArgument("kubelet", "--cluster-domain")
	if err != nil || clusterDomain != "cluster.local" {
		err := utils.UpdateServiceArgs("cluster-domain", "cluster.local", "kubelet")
		if err != nil {
			return fmt.Errorf("failed to update cluster-domain argument: %w", err)
		}
		err = utils.UpdateServiceArgs("cluster-dns", dnsIP, "kubelet")
		if err != nil {
			return fmt.Errorf("failed to update cluster-dns argument: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = utils.StopService(ctx, "kubelet")
		if err != nil {
			return fmt.Errorf("failed to stop service 'kubelet': %w", err)
		}
		err = utils.StartService(ctx, "kubelet")
		if err != nil {
			return fmt.Errorf("failed to start service 'kubelet': %w", err)
		}

	}

	return nil
}
