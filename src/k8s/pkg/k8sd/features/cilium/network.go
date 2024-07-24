package cilium

import (
	"context"
	"fmt"

	"github.com/canonical/k8s/pkg/client/helm"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/log"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/k8s/pkg/utils/control"
)

const (
	networkDeleteFailedMsgTmpl = "Failed to delete Cilium Network, the error was: %v"
	networkDeployFailedMsgTmpl = "Failed to deploy Cilium Network, the error was: %v"
)

// ApplyNetwork will deploy Cilium when cfg.Enabled is true.
// ApplyNetwork will remove Cilium when cfg.Enabled is false.
// ApplyNetwork requires that bpf and cgroups2 are already mounted and available when running under strict snap confinement. If they are not, it will fail (since Cilium will not have the required permissions to mount them).
// ApplyNetwork requires that `/sys` is mounted as a shared mount when running under classic snap confinement. This is to ensure that Cilium will be able to automatically mount bpf and cgroups2 on the pods.
// ApplyNetwork will always return a FeatureStatus indicating the current status of the
// deployment.
// ApplyNetwork returns an error if anything fails. The error is also wrapped in the .Message field of the
// returned FeatureStatus.
func ApplyNetwork(ctx context.Context, snap snap.Snap, cfg types.Network, _ types.Annotations) (types.FeatureStatus, error) {
	status := types.FeatureStatus{
		Version: ciliumAgentImageTag,
		Enabled: cfg.GetEnabled(),
	}
	m := snap.HelmClient()

	if !cfg.GetEnabled() {
		if _, err := m.Apply(ctx, chartCilium, helm.StateDeleted, nil); err != nil {
			uninstallErr := fmt.Errorf("failed to uninstall network: %w", err)
			status.Message = fmt.Sprintf(networkDeleteFailedMsgTmpl, uninstallErr)
			return status, uninstallErr
		}
		status.Message = disabledMsg
		status.Version = ""
		return status, nil
	}

	ipv4CIDR, ipv6CIDR, err := utils.ParseCIDRs(cfg.GetPodCIDR())
	if err != nil {
		cidrErr := fmt.Errorf("invalid kube-proxy --cluster-cidr value: %v", err)
		status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, cidrErr)
		status.Enabled = false
		return status, cidrErr
	}

	values := map[string]any{
		"image": map[string]any{
			"repository": ciliumAgentImageRepo,
			"tag":        ciliumAgentImageTag,
			"useDigest":  false,
		},
		"socketLB": map[string]any{
			"enabled": true,
		},
		"cni": map[string]any{
			"confPath": "/etc/cni/net.d",
			"binPath":  "/opt/cni/bin",
		},
		"operator": map[string]any{
			"replicas": 1,
			"image": map[string]any{
				"repository": ciliumOperatorImageRepo,
				"tag":        ciliumOperatorImageTag,
				"useDigest":  false,
			},
		},
		"ipv4": map[string]any{
			"enabled": ipv4CIDR != "",
		},
		"ipv6": map[string]any{
			"enabled": ipv6CIDR != "",
		},
		"ipam": map[string]any{
			"operator": map[string]any{
				"clusterPoolIPv4PodCIDRList": ipv4CIDR,
				"clusterPoolIPv6PodCIDRList": ipv6CIDR,
			},
		},
		"nodePort": map[string]any{
			"enabled": true,
		},
		"disableEnvoyVersionCheck": true,
	}

	if snap.Strict() {
		bpfMnt, err := utils.GetMountPath("bpf")
		if err != nil {
			mntErr := fmt.Errorf("failed to get bpf mount path: %w", err)
			status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, mntErr)
			status.Enabled = false
			return status, mntErr
		}

		cgrMnt, err := utils.GetMountPath("cgroup2")
		if err != nil {
			cgrpErr := fmt.Errorf("failed to get cgroup2 mount path: %w", err)
			status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, cgrpErr)
			status.Enabled = false
			return status, cgrpErr
		}

		values["bpf"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"root": bpfMnt,
		}
		values["cgroup"] = map[string]any{
			"autoMount": map[string]any{
				"enabled": false,
			},
			"hostRoot": cgrMnt,
		}
	} else {
		p, err := utils.GetMountPropagation("/sys")
		if err != nil {
			mntErr := fmt.Errorf("failed to get mount propagation for %s: %w", p, err)
			status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, mntErr)
			status.Enabled = false
			return status, mntErr
		}
		if p == "private" {
			onLXD, err := snap.OnLXD(ctx)
			if err != nil {
				log.FromContext(ctx).Error(err, "Failed to check if running on LXD")
			}
			if onLXD {
				lxdErr := fmt.Errorf("/sys is not a shared mount on the LXD container, this might be resolved by updating LXD on the host to version 5.0.2 or newer")
				status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, lxdErr)
				status.Enabled = false
				return status, lxdErr
			}

			sysErr := fmt.Errorf("/sys is not a shared mount")
			status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, sysErr)
			status.Enabled = false
			return status, sysErr
		}
	}

	if _, err := m.Apply(ctx, chartCilium, helm.StatePresent, values); err != nil {
		enableErr := fmt.Errorf("failed to enable network: %w", err)
		status.Message = fmt.Sprintf(networkDeployFailedMsgTmpl, enableErr)
		status.Enabled = false
		return status, enableErr
	}

	status.Message = enabledMsg
	return status, nil
}

func rolloutRestartCilium(ctx context.Context, snap snap.Snap, attempts int) error {
	client, err := snap.KubernetesClient("")
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDeployment(ctx, "cilium-operator", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium-operator deployment: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium-operator deployment after %d attempts: %w", attempts, err)
	}

	if err := control.RetryFor(ctx, attempts, 0, func() error {
		if err := client.RestartDaemonset(ctx, "cilium", "kube-system"); err != nil {
			return fmt.Errorf("failed to restart cilium daemonset: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to restart cilium daemonset after %d attempts: %w", attempts, err)
	}

	return nil
}
