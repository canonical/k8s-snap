package app

import (
	"fmt"
	"os"
	"path/filepath"

	apiv1 "github.com/canonical/k8s-snap-api/api/v1"
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
	secureBindAddress              = "127.0.0.1"

	disableAlphaAPIs = "AllAlpha=false"
)

func (a *App) ApplyComplianceProfile(profile string, joinConfig apiv1.WorkerJoinConfig, bootstrapConfig apiv1.BootstrapConfig, isWorker bool) error {
	switch profile {
	case ComplianceProfileDefault:
		if err := a.applyDefaultComplianceProfileRules(joinConfig, bootstrapConfig, isWorker); err != nil {
			return fmt.Errorf("failed to apply default compliance profile rules: %w", err)
		}
	case ComplianceProfileRecommended:
		if err := a.applyDefaultComplianceProfileRules(joinConfig, bootstrapConfig, isWorker); err != nil {
			return fmt.Errorf("failed to apply default compliance profile rules: %w", err)
		}
		if err := a.applyRecommendedComplianceProfileRules(joinConfig, bootstrapConfig, isWorker); err != nil {
			return fmt.Errorf("failed to apply recommended compliance profile rules: %w", err)
		}
	default:
		return fmt.Errorf("unknown compliance profile: %s", profile)
	}
	return nil
}

func (a *App) applyDefaultComplianceProfileRules(joinConfig apiv1.WorkerJoinConfig, bootstrapConfig apiv1.BootstrapConfig, isWorker bool) error {
	a.applyKubectlStreamingConnIdleTimeout(joinConfig, bootstrapConfig, isWorker)
	return nil
}

func (a *App) applyRecommendedComplianceProfileRules(joinConfig apiv1.WorkerJoinConfig, bootstrapConfig apiv1.BootstrapConfig, isWorker bool) error {
	if !isWorker {
		a.applySecureBindingKubeSchedulerControllerManager(bootstrapConfig)
		a.applyDisableAlphaAPIs(bootstrapConfig)
		if err := a.applyAuditLogging(bootstrapConfig); err != nil {
			return fmt.Errorf("failed to apply DISA STIG rules related to audit logging: %w", err)
		}
	}

	a.applyKubeletEnableKernelProtection(joinConfig, bootstrapConfig)
	return nil
}

// Rule V-242384, V-242385
func (a *App) applySecureBindingKubeSchedulerControllerManager(bootstrapConfig apiv1.BootstrapConfig) {
	// TODO: check IPv6
	if bootstrapConfig.ExtraNodeKubeSchedulerArgs == nil {
		bootstrapConfig.ExtraNodeKubeSchedulerArgs = make(map[string]*string)
	}
	if bootstrapConfig.ExtraNodeKubeControllerManagerArgs == nil {
		bootstrapConfig.ExtraNodeKubeControllerManagerArgs = make(map[string]*string)
	}
	bootstrapConfig.ExtraNodeKubeSchedulerArgs["--secure-bind-address"] = utils.Pointer(secureBindAddress)
	bootstrapConfig.ExtraNodeKubeControllerManagerArgs["--secure-bind-address"] = utils.Pointer(secureBindAddress)
}

// Rule V-242400
func (a *App) applyDisableAlphaAPIs(bootstrapConfig apiv1.BootstrapConfig) {
	if bootstrapConfig.ExtraNodeKubeAPIServerArgs == nil {
		bootstrapConfig.ExtraNodeKubeAPIServerArgs = make(map[string]*string)
	}
	bootstrapConfig.ExtraNodeKubeAPIServerArgs["--feature-gates"] = utils.Pointer(disableAlphaAPIs)
}

// Rules V-242402, V-242403, V-242461, V-242462, V-242463, V-242464, V-242465
func (a *App) applyAuditLogging(bootstrapConfig apiv1.BootstrapConfig) error {
	auditPolicyPath := filepath.Join(a.snap.EtcDir(), auditPolicyFileName)

	if err := os.MkdirAll(filepath.Dir(auditPolicyPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for audit policy: %w", err)
	}

	if err := utils.WriteFile(auditPolicyPath, []byte(auditPolicyTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write audit policy file: %w", err)
	}

	if bootstrapConfig.ExtraNodeKubeAPIServerArgs == nil {
		bootstrapConfig.ExtraNodeKubeAPIServerArgs = make(map[string]*string)
	}

	bootstrapConfig.ExtraNodeKubeAPIServerArgs["--audit-log-path"] = utils.Pointer(auditLogPath)
	bootstrapConfig.ExtraNodeKubeAPIServerArgs["--audit-policy-file"] = utils.Pointer(auditPolicyPath)
	bootstrapConfig.ExtraNodeKubeAPIServerArgs["--audit-log-maxage"] = utils.Pointer(auditLogMaxAge)
	bootstrapConfig.ExtraNodeKubeAPIServerArgs["--audit-log-maxbackup"] = utils.Pointer(auditLogMaxBackup)
	bootstrapConfig.ExtraNodeKubeAPIServerArgs["--audit-log-maxsize"] = utils.Pointer(auditLogMaxSize)

	return nil
}

// Rule V-242434
func (a *App) applyKubeletEnableKernelProtection(joinConfig apiv1.WorkerJoinConfig, bootstrapConfig apiv1.BootstrapConfig) error {
	// TODO: set sysctl values
	if joinConfig.ExtraNodeKubeletArgs == nil {
		joinConfig.ExtraNodeKubeletArgs = make(map[string]*string)
	}
	joinConfig.ExtraNodeKubeletArgs["--protect-kernel-defaults"] = utils.Pointer("true")
	return nil
}

// Rule V-245541
func (a *App) applyKubectlStreamingConnIdleTimeout(joinConfig apiv1.WorkerJoinConfig, bootstrapConfig apiv1.BootstrapConfig, isWorker bool) {
	if isWorker {
		if joinConfig.ExtraNodeKubeletArgs == nil {
			joinConfig.ExtraNodeKubeletArgs = make(map[string]*string)
		}
		joinConfig.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"] = utils.Pointer(streamingConnectionIdleTimeout)
	} else {
		if bootstrapConfig.ExtraNodeKubeletArgs == nil {
			bootstrapConfig.ExtraNodeKubeletArgs = make(map[string]*string)
		}
		bootstrapConfig.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"] = utils.Pointer(streamingConnectionIdleTimeout)
	}

}
