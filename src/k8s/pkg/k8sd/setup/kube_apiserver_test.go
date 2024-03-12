package setup_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestKubeAPIServer(t *testing.T) {
	g := NewWithT(t)

	dir := t.TempDir()

	s := &mock.Snap{
		Mock: mock.Mock{
			UID:                   os.Getuid(),
			GID:                   os.Getgid(),
			KubernetesConfigDir:   path.Join(dir, "kubernetes"),
			KubernetesPKIDir:      path.Join(dir, "kubernetes-pki"),
			ServiceArgumentsDir:   path.Join(dir, "args"),
			ServiceExtraConfigDir: path.Join(dir, "args/conf.d"),
			K8sDqliteStateDir:     path.Join(dir, "k8s-dqlite"),
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(BeNil())
	g.Expect(setup.KubeAPIServer(s, "10.152.0.0/16", "https://10.0.0.1:6400/1.0/kubernetes/auth/webhook", false, "k8s-dqlite", fmt.Sprintf("unix://%s", path.Join(s.K8sDqliteStateDir(), "k8s-dqlite.sock")), "Node,RBAC")).To(BeNil())
}
