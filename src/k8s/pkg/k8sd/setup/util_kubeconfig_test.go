package setup_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	. "github.com/onsi/gomega"
)

func TestKubeconfigString(t *testing.T) {
	g := NewWithT(t)

	ca := base64.StdEncoding.EncodeToString([]byte("ca"))
	expectedConfig := fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: %s
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
    token: token
`, ca)

	actual, _ := setup.KubeconfigString("token", "server", "ca")

	g.Expect(actual).To(Equal(expectedConfig))
}
