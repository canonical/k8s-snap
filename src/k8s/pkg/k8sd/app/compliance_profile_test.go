package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/app"
	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestApplyComplianceProfile(t *testing.T) {
	g := NewWithT(t)
	tempDir := t.TempDir()

	app := &app.App{}

	// Setup mock snap
	snapMock := &mock.Snap{
		Mock: mock.Mock{
			EtcDir: tempDir,
		},
	}

	t.Run("DefaultProfile", func(t *testing.T) {
		serviceConfigs := &types.K8sServiceConfigs{
			ExtraNodeKubeletArgs: make(map[string]*string),
		}

		err := app.ApplyComplianceProfile("default", serviceConfigs, snapMock, false)
		g.Expect(err).NotTo(HaveOccurred())

		// Check that streaming connection timeout is applied
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).To(HaveKey("--streaming-connection-idle-timeout"))
		g.Expect(*serviceConfigs.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"]).To(Equal("5m"))
	})

	t.Run("RecommendedProfileControlPlane", func(t *testing.T) {
		serviceConfigs := &types.K8sServiceConfigs{
			ExtraNodeKubeletArgs:               make(map[string]*string),
			ExtraNodeKubeSchedulerArgs:         make(map[string]*string),
			ExtraNodeKubeControllerManagerArgs: make(map[string]*string),
			ExtraNodeKubeAPIServerArgs:         make(map[string]*string),
		}

		err := app.ApplyComplianceProfile("recommended", serviceConfigs, snapMock, true)
		g.Expect(err).NotTo(HaveOccurred())

		// Check kubelet args
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).To(HaveKey("--streaming-connection-idle-timeout"))
		g.Expect(*serviceConfigs.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"]).To(Equal("5m"))
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).To(HaveKey("--protect-kernel-defaults"))
		g.Expect(*serviceConfigs.ExtraNodeKubeletArgs["--protect-kernel-defaults"]).To(Equal("true"))

		// Check scheduler secure binding
		g.Expect(serviceConfigs.ExtraNodeKubeSchedulerArgs).To(HaveKey("--secure-bind-address"))
		g.Expect(*serviceConfigs.ExtraNodeKubeSchedulerArgs["--secure-bind-address"]).To(Equal("127.0.0.1"))

		// Check controller manager secure binding
		g.Expect(serviceConfigs.ExtraNodeKubeControllerManagerArgs).To(HaveKey("--secure-bind-address"))
		g.Expect(*serviceConfigs.ExtraNodeKubeControllerManagerArgs["--secure-bind-address"]).To(Equal("127.0.0.1"))

		// Check API server args
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--feature-gates"))
		g.Expect(*serviceConfigs.ExtraNodeKubeAPIServerArgs["--feature-gates"]).To(Equal("AllAlpha=false"))

		// Check audit logging
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--audit-log-path"))
		g.Expect(*serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-path"]).To(Equal("/var/log/kubernetes/audit.log"))
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--audit-policy-file"))
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--audit-log-maxage"))
		g.Expect(*serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-maxage"]).To(Equal("30"))
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--audit-log-maxbackup"))
		g.Expect(*serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-maxbackup"]).To(Equal("10"))
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--audit-log-maxsize"))
		g.Expect(*serviceConfigs.ExtraNodeKubeAPIServerArgs["--audit-log-maxsize"]).To(Equal("100"))
	})

	t.Run("RecommendedProfileWorker", func(t *testing.T) {
		serviceConfigs := &types.K8sServiceConfigs{
			ExtraNodeKubeletArgs: make(map[string]*string),
		}

		err := app.ApplyComplianceProfile("recommended", serviceConfigs, snapMock, false)
		g.Expect(err).NotTo(HaveOccurred())

		// Worker nodes should only get kubelet config, no control plane components
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).To(HaveKey("--streaming-connection-idle-timeout"))
		g.Expect(*serviceConfigs.ExtraNodeKubeletArgs["--streaming-connection-idle-timeout"]).To(Equal("5m"))
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).To(HaveKey("--protect-kernel-defaults"))
		g.Expect(*serviceConfigs.ExtraNodeKubeletArgs["--protect-kernel-defaults"]).To(Equal("true"))

		// Control plane args should not be set for worker nodes
		g.Expect(serviceConfigs.ExtraNodeKubeSchedulerArgs).To(BeEmpty())
		g.Expect(serviceConfigs.ExtraNodeKubeControllerManagerArgs).To(BeEmpty())
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(BeEmpty())
	})

	t.Run("UnknownProfile", func(t *testing.T) {
		serviceConfigs := &types.K8sServiceConfigs{}

		err := app.ApplyComplianceProfile("unknown", serviceConfigs, snapMock, false)
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring("unknown compliance profile"))
	})

	t.Run("NilMapsHandling", func(t *testing.T) {
		// Test that nil maps are handled gracefully
		serviceConfigs := &types.K8sServiceConfigs{
			// All maps are nil initially
		}

		err := app.ApplyComplianceProfile("recommended", serviceConfigs, snapMock, true)
		g.Expect(err).NotTo(HaveOccurred())

		// Maps should be created and populated
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).NotTo(BeNil())
		g.Expect(serviceConfigs.ExtraNodeKubeletArgs).To(HaveKey("--streaming-connection-idle-timeout"))

		g.Expect(serviceConfigs.ExtraNodeKubeSchedulerArgs).NotTo(BeNil())
		g.Expect(serviceConfigs.ExtraNodeKubeSchedulerArgs).To(HaveKey("--secure-bind-address"))

		g.Expect(serviceConfigs.ExtraNodeKubeControllerManagerArgs).NotTo(BeNil())
		g.Expect(serviceConfigs.ExtraNodeKubeControllerManagerArgs).To(HaveKey("--secure-bind-address"))

		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).NotTo(BeNil())
		g.Expect(serviceConfigs.ExtraNodeKubeAPIServerArgs).To(HaveKey("--feature-gates"))
	})

	t.Run("AuditPolicyFileCreated", func(t *testing.T) {
		serviceConfigs := &types.K8sServiceConfigs{
			ExtraNodeKubeAPIServerArgs: make(map[string]*string),
		}

		err := app.ApplyComplianceProfile("recommended", serviceConfigs, snapMock, true)
		g.Expect(err).NotTo(HaveOccurred())

		// Check that audit policy file was created
		auditPolicyPath := filepath.Join(tempDir, "audit-policy.yaml")
		g.Expect(auditPolicyPath).To(BeAnExistingFile())

		// Check that the audit policy file contains expected content
		content, err := os.ReadFile(auditPolicyPath)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(string(content)).To(ContainSubstring("apiVersion: audit.k8s.io/v1"))
	})
}
