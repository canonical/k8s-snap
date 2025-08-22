package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
)

const (
	ComplianceProfileDefault     = "default"
	ComplianceProfileRecommended = "recommended"

	auditLogPath        = "/var/log/kubernetes/audit.log"
	auditPolicyFileName = "audit-policy.yaml"
	auditPolicyTemplate = `# Log all requests at the Metadata level.
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: Metadata
`
	auditLogMaxAge    = "30"
	auditLogMaxBackup = "10"
	auditLogMaxSize   = "100"

	streamingConnectionIdleTimeout = "5m"

	disableAlphaAPIs = "AllAlpha=false"
)

// ApplyComplianceProfile applies the specified compliance profile to the given service configurations.
func (a *App) ApplyComplianceProfile(profile string, serviceConfigs *types.K8sServiceConfigs, s snap.Snap, isControlPlane bool) error {
	switch profile {
	case ComplianceProfileDefault:
		if err := a.applyDefaultComplianceProfileRules(serviceConfigs); err != nil {
			return fmt.Errorf("failed to apply default compliance profile rules: %w", err)
		}
	case ComplianceProfileRecommended:
		if err := a.applyDefaultComplianceProfileRules(serviceConfigs); err != nil {
			return fmt.Errorf("failed to apply default compliance profile rules: %w", err)
		}
		if err := a.applyRecommendedComplianceProfileRules(serviceConfigs, s, isControlPlane); err != nil {
			return fmt.Errorf("failed to apply recommended compliance profile rules: %w", err)
		}
	default:
		return fmt.Errorf("unknown compliance profile: %s", profile)
	}
	return nil
}

func (a *App) applyDefaultComplianceProfileRules(serviceConfigs *types.K8sServiceConfigs) error {
	a.applyKubectlStreamingConnIdleTimeout(serviceConfigs)
	return nil
}

func (a *App) applyRecommendedComplianceProfileRules(serviceConfigs *types.K8sServiceConfigs, s snap.Snap, isControlPlane bool) error {
	if isControlPlane {
		if err := a.applySecureBindingKubeSchedulerControllerManager(serviceConfigs); err != nil {
			return fmt.Errorf("failed to apply secure binding for kube-scheduler and kube-controller-manager: %w", err)
		}
		a.applyDisableAlphaAPIs(serviceConfigs)
		if err := a.applyAuditLogging(serviceConfigs, s); err != nil {
			return fmt.Errorf("failed to apply DISA STIG rules related to audit logging: %w", err)
		}
	}

	a.applyKubeletEnableKernelProtection(serviceConfigs)
	return nil
}

// Rule V-242384, V-242385.
func (a *App) applySecureBindingKubeSchedulerControllerManager(serviceConfigs *types.K8sServiceConfigs) error {
	if serviceConfigs.ExtraNodeKubeSchedulerArgs == nil {
		serviceConfigs.ExtraNodeKubeSchedulerArgs = make(map[string]*string)
	}
	if serviceConfigs.ExtraNodeKubeControllerManagerArgs == nil {
		serviceConfigs.ExtraNodeKubeControllerManagerArgs = make(map[string]*string)
	}
	secureBindAddress, err := utils.GetLocalhostAddress()
	if err != nil {
		return fmt.Errorf("Failed to get localhost address: %w", err)
	}
	serviceConfigs.ExtraNodeKubeSchedulerArgs["--secure-bind-address"] = utils.Pointer(secureBindAddress.String())
	serviceConfigs.ExtraNodeKubeControllerManagerArgs["--secure-bind-address"] = utils.Pointer(secureBindAddress.String())
	return nil
}

// Rule V-242400.
func (a *App) applyDisableAlphaAPIs(serviceConfigs *types.K8sServiceConfigs) {
	if serviceConfigs.ExtraNodeKubeAPIServerArgs == nil {
		serviceConfigs.ExtraNodeKubeAPIServerArgs = make(map[string]*string)
	}
	serviceConfigs.ExtraNodeKubeAPIServerArgs["--feature-gates"] = utils.Pointer(disableAlphaAPIs)
}

// Rules V-242402, V-242403, V-242461, V-242462, V-242463, V-242464, V-242465.
func (a *App) applyAuditLogging(serviceConfigs *types.K8sServiceConfigs, s snap.Snap) error {
	auditPolicyPath := filepath.Join(s.EtcDir(), auditPolicyFileName)

	if err := utils.WriteFile(auditPolicyPath, []byte(auditPolicyTemplate), os.FileMode(0o644)); err != nil {
		return fmt.Errorf("failed to write audit policy file: %w", err)
	}

	if serviceConfigs.ExtraNodeKubeAPIServerArgs == nil {
		serviceConfigs.ExtraNodeKubeAPIServerArgs = make(map[string]*string)
	}

	serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-path"] = utils.Pointer(auditLogPath)
	serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-policy-file"] = utils.Pointer(auditPolicyPath)
	serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-maxage"] = utils.Pointer(auditLogMaxAge)
	serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-maxbackup"] = utils.Pointer(auditLogMaxBackup)
	serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-maxsize"] = utils.Pointer(auditLogMaxSize)

	return nil
}

// Rule V-242434.
func (a *App) applyKubeletEnableKernelProtection(serviceConfigs *types.K8sServiceConfigs) error {
	if serviceConfigs.ExtraNodeKubeletArgs == nil {
		serviceConfigs.ExtraNodeKubeletArgs = make(map[string]*string)
	}
	serviceConfigs.ExtraNodeKubeletArgs["--protect-kernel-defaults"] = utils.Pointer("true")
	return nil
}

// Rule V-245541.
func (a *App) applyKubectlStreamingConnIdleTimeout(serviceConfigs *types.K8sServiceConfigs) {
	if serviceConfigs.ExtraNodeKubeletArgs == nil {
		serviceConfigs.ExtraNodeKubeletArgs = make(map[string]*string)
	}
	serviceConfigs.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"] = utils.Pointer(streamingConnectionIdleTimeout)
}
