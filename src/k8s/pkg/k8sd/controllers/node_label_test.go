package controllers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/canonical/k8s/pkg/client/kubernetes"
	"github.com/canonical/k8s/pkg/k8sd/setup"
	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestAvailabilityZoneLabel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g := NewWithT(t)

	tests := []struct {
		name string
		// Availability zone label
		availabilityZone string
		// The first 8 bytes of the sha-256 hash of the AZ or 0 if unset.
		expFailureDomain uint64
		// Does the $stateDir/failure-domain file exist?
		fileExists bool
		// The already existing failure domain setting.
		existingFailureDomain uint64
		// Do we expect a service restart?
		expRestart bool
	}{
		{
			name:             "AZ set, file missing",
			availabilityZone: "testAZ",
			expFailureDomain: 7130520900010879344,
			fileExists:       false,
			expRestart:       true,
		},
		{
			name:                  "No change - file missing",
			availabilityZone:      "",
			expFailureDomain:      0,
			existingFailureDomain: 0,
			fileExists:            false,
			expRestart:            false,
		},
		{
			name:                  "No change, AZ unset - file exists",
			availabilityZone:      "",
			expFailureDomain:      0,
			existingFailureDomain: 0,
			fileExists:            true,
			expRestart:            false,
		},
		{
			name:                  "No change, AZ set - file exists",
			availabilityZone:      "testAZ",
			expFailureDomain:      7130520900010879344,
			existingFailureDomain: 7130520900010879344,
			fileExists:            true,
			expRestart:            false,
		},
		{
			name:                  "AZ changed, file exists",
			availabilityZone:      "testAZ",
			expFailureDomain:      7130520900010879344,
			existingFailureDomain: 101,
			fileExists:            true,
			expRestart:            true,
		},
	}

	clientset := fake.NewSimpleClientset()
	watcher := watch.NewFake()
	clientset.PrependWatchReactor("nodes", k8stesting.DefaultWatchReactor(watcher, nil))

	s := &mock.Snap{
		Mock: mock.Mock{
			K8sdStateDir:         filepath.Join(t.TempDir(), "k8sd"),
			K8sDqliteStateDir:    filepath.Join(t.TempDir(), "k8s-dqlite"),
			UID:                  os.Getuid(),
			GID:                  os.Getgid(),
			KubernetesNodeClient: &kubernetes.Client{Interface: clientset},
		},
	}

	g.Expect(setup.EnsureAllDirectories(s)).To(Succeed())

	k8sDqliteStateDir := s.K8sDqliteStateDir()
	k8sdDbDir := filepath.Join(s.K8sdStateDir(), "database")

	// EnsureAllDirectories doesn't handle the following k8sd dirs, for
	// now we'll create them here.
	g.Expect(os.MkdirAll(k8sDqliteStateDir, 0o700)).To(Succeed())
	g.Expect(os.MkdirAll(k8sdDbDir, 0o700)).To(Succeed())

	ctrl := NewNodeLabelController(s, func() {})

	go ctrl.Run(ctx)
	defer watcher.Stop()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s.RestartServiceCalledWith = nil

			k8sDqliteFailureDomainFile := snaputil.GetDqliteFailureDomainFile(k8sDqliteStateDir)
			k8sdFailureDomainFile := snaputil.GetDqliteFailureDomainFile(k8sdDbDir)

			if tc.fileExists {
				existingFailureDomainStr := fmt.Sprintf("%v", tc.existingFailureDomain)
				// For simplicity, we'll assume matching dqlite failure domains
				err := os.WriteFile(k8sDqliteFailureDomainFile, []byte(existingFailureDomainStr), 0o644)
				g.Expect(err).ToNot(HaveOccurred())
				err = os.WriteFile(k8sdFailureDomainFile, []byte(existingFailureDomainStr), 0o644)
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				exists, err := utils.FileExists(k8sDqliteFailureDomainFile)
				g.Expect(err).ToNot(HaveOccurred())
				if exists {
					g.Expect(os.Remove(k8sDqliteFailureDomainFile)).To(Succeed())
				}
				exists, err = utils.FileExists(k8sdFailureDomainFile)
				g.Expect(err).ToNot(HaveOccurred())
				if exists {
					g.Expect(os.Remove(k8sdFailureDomainFile)).To(Succeed())
				}
			}

			g := NewWithT(t)

			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: s.Hostname(),
					Labels: map[string]string{
						"topology.kubernetes.io/zone": tc.availabilityZone,
					},
				},
			}
			watcher.Add(node)

			// TODO: this is to ensure that the controller has handled the event. This should ideally
			// be replaced with something like a "<-sentCh" instead
			time.Sleep(100 * time.Millisecond)

			k8sdFailureDomain, err := snaputil.GetDqliteFailureDomain(k8sdDbDir)
			g.Expect(err).ToNot(HaveOccurred())
			k8sdDqliteFailureDomain, err := snaputil.GetDqliteFailureDomain(k8sDqliteStateDir)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(k8sdFailureDomain).To(Equal(k8sdDqliteFailureDomain))
			g.Expect(k8sdFailureDomain).To(Equal(tc.expFailureDomain))

			if tc.expRestart {
				g.Expect(s.RestartServiceCalledWith).To(ContainElement("k8sd"))
				g.Expect(s.RestartServiceCalledWith).To(ContainElement("k8s-dqlite"))
			} else {
				g.Expect(s.RestartServiceCalledWith).To(BeEmpty())
			}
		})
	}
}
