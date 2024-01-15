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

	err = utils.UpdateServiceArgs("cluster-dns", dnsIP, "kubelet")
	if err != nil {
		return fmt.Errorf("failed to update cluster-dns argument: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = utils.StopService(ctx, "kubelet")
	if err != nil {
		return fmt.Errorf("failed to stop service 'kubelet': %w", err)
	}
	err = utils.StartService(ctx, "kubelet")
	if err != nil {
		return fmt.Errorf("failed to start service 'kubelet': %w", err)
	}

	return nil
}
