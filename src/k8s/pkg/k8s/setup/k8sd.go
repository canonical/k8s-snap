package setup

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/canonical/k8s/pkg/k8s/client"
)

// InitK8sd handles the setup of K8sd.
func InitK8sd(ctx context.Context, clusterOpts client.ClusterOpts) (*client.Client, error) {
	startCmd := exec.Command("snapctl", "start", "k8s.k8sd")
	var err error

	_, err = startCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}

	var cl *client.Client
	// var member v1.ClusterMember

	ch := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(2 * time.Second)
			cl, err = client.NewClient(ctx, clusterOpts)
			if err != nil {
				err = fmt.Errorf("failed to create client: %w", err)
				continue
			}

			// TODO (KU-166): It's not yet possible to join two bootstrapped clusters
			// member, err = cl.Bootstrap(ctx)
			// if err != nil {
			// 	err = fmt.Errorf("failed to bootstrap cluster: %w", err)
			// 	continue
			// }
			break
		}
		ch <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		// logrus.Infof("Cluster with member %s on %s created.", member.Name, member.Address)
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timed out while waiting for k8sd initialization: %w", err)
	}

	return cl, nil
}
