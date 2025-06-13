package k8s

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"gopkg.in/yaml.v2"
)

type runCommand func(ctx context.Context, command []string, opts ...func(c *exec.Cmd)) error

func verifyBootstrapConfig(bootstrapConfig apiv1.BootstrapConfig) error {
	return verifyBootstrapConfigWithRunCommand(bootstrapConfig, nil)
}

func verifyBootstrapConfigWithRunCommand(bootstrapConfig apiv1.BootstrapConfig, run runCommand) error {
	cfg, err := types.ClusterConfigFromBootstrapConfig(bootstrapConfig)
	if err != nil {
		return err
	}

	cfg.SetDefaults()

	svcConfigs := types.K8sServiceConfigs{
		ExtraNodeKubeSchedulerArgs:         bootstrapConfig.ExtraNodeKubeSchedulerArgs,
		ExtraNodeKubeControllerManagerArgs: bootstrapConfig.ExtraNodeKubeControllerManagerArgs,
		ExtraNodeKubeletArgs:               bootstrapConfig.ExtraNodeKubeletArgs,
		ExtraNodeKubeProxyArgs:             bootstrapConfig.ExtraNodeKubeProxyArgs,
	}

	return verifyConfig(cfg, svcConfigs, bootstrapConfig.ContainerdBaseDir, true, run)
}

func verifyJoinConfig(joinConfigString, token string) error {
	return verifyJoinConfigWithRunCommand(joinConfigString, token, nil)
}

func verifyJoinConfigWithRunCommand(joinConfigString, token string, run runCommand) error {
	cfg := types.ClusterConfig{}
	cfg.SetDefaults()

	internalToken := types.InternalWorkerNodeToken{}
	if internalToken.Decode(token) == nil {
		// worker token.
		var joinConfig apiv1.WorkerJoinConfig
		if err := yaml.UnmarshalStrict([]byte(joinConfigString), &joinConfig); err != nil {
			return fmt.Errorf("failed to unmarshal worker join config: %w", err)
		}

		svcConfigs := types.K8sServiceConfigs{
			ExtraNodeKubeletArgs:   joinConfig.ExtraNodeKubeletArgs,
			ExtraNodeKubeProxyArgs: joinConfig.ExtraNodeKubeProxyArgs,
		}

		return verifyConfig(cfg, svcConfigs, joinConfig.ContainerdBaseDir, false, run)
	}

	var joinConfig apiv1.ControlPlaneJoinConfig
	if err := yaml.UnmarshalStrict([]byte(joinConfigString), &joinConfig); err != nil {
		return fmt.Errorf("failed to unmarshal control plane join config: %w", err)
	}

	svcConfigs := types.K8sServiceConfigs{
		ExtraNodeKubeSchedulerArgs:         joinConfig.ExtraNodeKubeSchedulerArgs,
		ExtraNodeKubeControllerManagerArgs: joinConfig.ExtraNodeKubeControllerManagerArgs,
		ExtraNodeKubeletArgs:               joinConfig.ExtraNodeKubeletArgs,
		ExtraNodeKubeProxyArgs:             joinConfig.ExtraNodeKubeProxyArgs,
	}

	return verifyConfig(cfg, svcConfigs, joinConfig.ContainerdBaseDir, true, run)
}

func verifyConfig(cfg types.ClusterConfig, svcConfigs types.K8sServiceConfigs, containerdBaseDir string, isControlPlane bool, run runCommand) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid cluster configuration: %w", err)
	}

	s := snap.NewSnap(snap.SnapOpts{
		SnapDir:           os.Getenv("SNAP"),
		SnapCommonDir:     os.Getenv("SNAP_COMMON"),
		SnapInstanceName:  os.Getenv("SNAP_INSTANCE_NAME"),
		ContainerdBaseDir: containerdBaseDir,
		RunCommand:        run,
	})

	// Pre-init checks
	if err := s.PreInitChecks(context.Background(), cfg, svcConfigs, isControlPlane); err != nil {
		return fmt.Errorf("pre-init checks failed for node: %w", err)
	}

	return nil
}
