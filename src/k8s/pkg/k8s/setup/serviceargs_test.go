package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestInitServiceArgs(t *testing.T) {
	snapDir := t.TempDir()
	snapDataDir := t.TempDir()
	snapArgsDir := filepath.Join(snapDir, "k8s/args")
	dataArgsDir := filepath.Join(snapDataDir, "args")

	// Replace the snap instance with the temporary directory for testing
	mockSnap := snap.NewSnap(
		snapDir,
		snapDataDir,
		"",
	)

	testCases := []struct {
		name                 string
		initialFileContents  map[string]string
		overwrites           map[string]map[string]string
		expectedFileContents map[string]string
	}{
		{
			name: "joinOverwrites",
			initialFileContents: map[string]string{
				"kube-apiserver": `--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key
--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP
--secure-port=6443
--service-cluster-ip-range=10.152.183.0/24
`,
			},
			overwrites: map[string]map[string]string{
				"kube-apiserver": {
					"--secure-port": "6000",
				},
			},
			expectedFileContents: map[string]string{
				"kube-apiserver": `--kubelet-client-key=/etc/kubernetes/pki/apiserver-kubelet-client.key
--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP
--secure-port=6000
--service-cluster-ip-range=10.152.183.0/24
`,
			},
		},
		{
			name: "emptyOverwrites",
			initialFileContents: map[string]string{
				"kube-apiserver": `--authorization-mode=Node,RBAC
--client-ca-file=/etc/kubernetes/pki/ca.crt
--service-account-key-file=/etc/kubernetes/pki/serviceaccount.key
--service-cluster-ip-range=10.152.183.0/24
--tls-cert-file=/etc/kubernetes/pki/apiserver.crt
--tls-private-key-file=/etc/kubernetes/pki/apiserver.key
`,
			},
			overwrites: map[string]map[string]string{},
			expectedFileContents: map[string]string{
				"kube-apiserver": `--authorization-mode=Node,RBAC
--client-ca-file=/etc/kubernetes/pki/ca.crt
--service-account-key-file=/etc/kubernetes/pki/serviceaccount.key
--service-cluster-ip-range=10.152.183.0/24
--tls-cert-file=/etc/kubernetes/pki/apiserver.crt
--tls-private-key-file=/etc/kubernetes/pki/apiserver.key
`,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			mustSetupArgsDirectoriesAndFiles(t, snapArgsDir, dataArgsDir, tc.initialFileContents)

			err := InitServiceArgs(mockSnap, tc.overwrites)
			g.Expect(err).NotTo(HaveOccurred())

			// Verify the content of the argument files in the temporary directory
			verifyArgumentFileContent(t, dataArgsDir, tc.expectedFileContents)
		})
	}
}

func mustSetupArgsDirectoriesAndFiles(t *testing.T, snapArgsDir string, dataArgsDir string, fileContents map[string]string) {
	g := NewWithT(t)
	err := os.MkdirAll(snapArgsDir, 0755)
	g.Expect(err).To(BeNil())

	err = os.MkdirAll(dataArgsDir, 0755)
	g.Expect(err).To(BeNil())

	for _, service := range k8sServices {
		content, exists := fileContents[service]
		if !exists {
			content = ""
		}
		err := os.WriteFile(filepath.Join(snapArgsDir, service), []byte(content), 0755)
		g.Expect(err).To(BeNil())
	}
}

func verifyArgumentFileContent(t *testing.T, dataArgsDir string, expected map[string]string) {
	g := NewWithT(t)
	for service, expectedContent := range expected {
		filePath := filepath.Join(dataArgsDir, service)

		content, err := utils.ReadFile(filePath)
		g.Expect(err).NotTo(HaveOccurred())

		g.Expect(content).To(Equal(expectedContent))
	}
}
