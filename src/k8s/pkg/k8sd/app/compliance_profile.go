package app

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8sd/types"
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

func (a *App) ApplyComplianceProfile(profile string, serviceConfigs *types.K8sServiceConfigs, nodeIPs []net.IP, isControlPlane bool) error {
	switch profile {
	case ComplianceProfileDefault:
		if err := a.applyDefaultComplianceProfileRules(serviceConfigs); err != nil {
			return fmt.Errorf("failed to apply default compliance profile rules: %w", err)
		}
	case ComplianceProfileRecommended:
		if err := a.applyDefaultComplianceProfileRules(serviceConfigs); err != nil {
			return fmt.Errorf("failed to apply default compliance profile rules: %w", err)
		}
		if err := a.applyRecommendedComplianceProfileRules(serviceConfigs, nodeIPs, isControlPlane); err != nil {
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

func (a *App) applyRecommendedComplianceProfileRules(serviceConfigs *types.K8sServiceConfigs, nodeIPs []net.IP, isControlPlane bool) error {
	if isControlPlane {
		a.applySecureBindingKubeSchedulerControllerManager(serviceConfigs, nodeIPs)
		a.applyDisableAlphaAPIs(serviceConfigs)
		if err := a.applyAuditLogging(serviceConfigs); err != nil {
			return fmt.Errorf("failed to apply DISA STIG rules related to audit logging: %w", err)
		}
	}

	a.applyKubeletEnableKernelProtection(serviceConfigs)
	return nil
}

// Rule V-242384, V-242385
func (a *App) applySecureBindingKubeSchedulerControllerManager(serviceConfigs *types.K8sServiceConfigs, nodeIPs []net.IP) {
	if serviceConfigs.ExtraNodeKubeSchedulerArgs == nil {
		serviceConfigs.ExtraNodeKubeSchedulerArgs = make(map[string]*string)
	}
	if serviceConfigs.ExtraNodeKubeControllerManagerArgs == nil {
		serviceConfigs.ExtraNodeKubeControllerManagerArgs = make(map[string]*string)
	}
	secureBindAddress := getSecureBindAddress(nodeIPs)
	serviceConfigs.ExtraNodeKubeSchedulerArgs["--secure-bind-address"] = utils.Pointer(secureBindAddress)
	serviceConfigs.ExtraNodeKubeControllerManagerArgs["--secure-bind-address"] = utils.Pointer(secureBindAddress)
}

// Rule V-242400
func (a *App) applyDisableAlphaAPIs(serviceConfigs *types.K8sServiceConfigs) {
	if serviceConfigs.ExtraNodeKubeAPIServerArgs == nil {
		serviceConfigs.ExtraNodeKubeAPIServerArgs = make(map[string]*string)
	}
	serviceConfigs.ExtraNodeKubeAPIServerArgs["--feature-gates"] = utils.Pointer(disableAlphaAPIs)
}

// Rules V-242402, V-242403, V-242461, V-242462, V-242463, V-242464, V-242465
func (a *App) applyAuditLogging(serviceConfigs *types.K8sServiceConfigs) error {
	auditPolicyPath := filepath.Join(a.snap.EtcDir(), auditPolicyFileName)

	if err := os.MkdirAll(filepath.Dir(auditPolicyPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for audit policy: %w", err)
	}

	if err := utils.WriteFile(auditPolicyPath, []byte(auditPolicyTemplate), 0644); err != nil {
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

// Rule V-242434
func (a *App) applyKubeletEnableKernelProtection(serviceConfigs *types.K8sServiceConfigs) error {
	// TODO: set sysctl values
	if serviceConfigs.ExtraNodeKubeletArgs == nil {
		serviceConfigs.ExtraNodeKubeletArgs = make(map[string]*string)
	}
	serviceConfigs.ExtraNodeKubeletArgs["--protect-kernel-defaults"] = utils.Pointer("true")
	return nil
}

// Rule V-245541
func (a *App) applyKubectlStreamingConnIdleTimeout(serviceConfigs *types.K8sServiceConfigs) {
	if serviceConfigs.ExtraNodeKubeletArgs == nil {
		serviceConfigs.ExtraNodeKubeletArgs = make(map[string]*string)
	}
	serviceConfigs.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"] = utils.Pointer(streamingConnectionIdleTimeout)
}

func getSecureBindAddress(nodeIPs []net.IP) string {
	hasIPv6 := false
	hasIPv4 := false

	for _, ip := range nodeIPs {
		if ip.To4() != nil {
			hasIPv4 = true
		} else {
			hasIPv6 = true
		}
	}

	switch {
	case hasIPv6 && hasIPv4:
		return "127.0.0.1"
	case hasIPv6:
		return "::1"
	default:
		return "127.0.0.1"
	}
}
