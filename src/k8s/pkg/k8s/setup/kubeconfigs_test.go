package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestRenderKubeconfig(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// Use the repository dir as the snap dir so that the template path is resolved
	snapDir := filepath.Join(wd, "../../../../../")
	mockSnap := snap.NewSnap(
		snapDir,
		"",
		"",
	)

	testCases := []struct {
		name               string
		hostOverwrite      string
		portOverwrite      string
		expectedKubeconfig string
	}{
		{
			name:          "withOverwrites",
			hostOverwrite: "192.168.12.3",
			portOverwrite: "6000",
			expectedKubeconfig: `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: bW9ja1BlbQ==
    server: https://{{ .ApiServerIp }}:{{ .ApiServerPort }}
  name: k8s
contexts:
- context:
    cluster: k8s
    user: k8s-user
  name: k8s
current-context: k8s
kind: Config
users:
- name: k8s-user
  user:
    token: token
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			file := filepath.Join(t.TempDir(), tc.name)
			err := renderKubeconfig(mockSnap, "token", []byte("mockPem"), file, &tc.hostOverwrite, &tc.portOverwrite)
			g.Expect(err).NotTo(HaveOccurred())
			content, err := utils.ReadFile(file)
			g.Expect(err).NotTo(HaveOccurred())

			expectedKubeconfig := strings.ReplaceAll(tc.expectedKubeconfig, "{{ .ApiServerIp }}", tc.hostOverwrite)
			expectedKubeconfig = strings.ReplaceAll(expectedKubeconfig, "{{ .ApiServerPort }}", tc.portOverwrite)
			g.Expect(content).To(Equal(expectedKubeconfig))
		})
	}
}
