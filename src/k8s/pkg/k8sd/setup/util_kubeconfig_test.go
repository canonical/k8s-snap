package setup_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	. "github.com/onsi/gomega"
)

func TestKubeconfigString(t *testing.T) {
	g := NewWithT(t)

	expectedConfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: Y2E=
    server: https://server
  name: k8s
contexts:
- context:
    cluster: k8s
    user: k8s-user
  name: k8s
current-context: k8s
kind: Config
preferences: {}
users:
- name: k8s-user
  user:
    client-certificate-data: Y3J0
    client-key-data: a2V5
`

	actual, err := setup.KubeconfigString("server", "ca", "crt", "key")

	g.Expect(actual).To(Equal(expectedConfig))
	g.Expect(err).To(BeNil())
}
