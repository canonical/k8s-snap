package snap_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestPebble(t *testing.T) {
	t.Run("Start", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.StartServices(context.Background(), []string{"test-service"})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble start test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartServices(context.Background(), []string{"test-service"})
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Stop", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})
		err := snap.StopServices(context.Background(), []string{"test-service"})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble stop test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartServices(context.Background(), []string{"test-service"})
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Restart", func(t *testing.T) {
		g := NewWithT(t)
		mockRunner := &mock.Runner{}
		snap := snap.NewPebble(snap.PebbleOpts{
			SnapDir:       "testdir",
			SnapCommonDir: "testdir",
			RunCommand:    mockRunner.Run,
		})

		err := snap.RestartServices(context.Background(), []string{"test-service"})
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(mockRunner.CalledWithCommand).To(ConsistOf("testdir/bin/pebble restart test-service"))

		t.Run("Fail", func(t *testing.T) {
			g := NewWithT(t)
			mockRunner.Err = fmt.Errorf("some error")

			err := snap.StartServices(context.Background(), []string{"service"})
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Revision", func(t *testing.T) {
		t.Run("returns revision from bom.json", func(t *testing.T) {
			g := NewWithT(t)

			tmpDir, err := os.MkdirTemp("", "test-bom-k8s")
			g.Expect(err).To(Not(HaveOccurred()))
			defer os.RemoveAll(tmpDir)

			revision := "f001154"
			bomContent := fmt.Sprintf(`{
				"k8s": {
					"revision": "%s"
				}
			}`, revision)

			err = os.WriteFile(filepath.Join(tmpDir, "bom.json"), []byte(bomContent), 0o644)
			g.Expect(err).To(Not(HaveOccurred()))

			snap := snap.NewPebble(snap.PebbleOpts{
				SnapDir: tmpDir,
			})

			returned, err := snap.Revision(context.Background())
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(returned).To(Equal(revision))
		})

		t.Run("fails when bom.json is missing", func(t *testing.T) {
			g := NewWithT(t)

			tmpDir, err := os.MkdirTemp("", "test-bom-missing")
			g.Expect(err).To(Not(HaveOccurred()))
			defer os.RemoveAll(tmpDir)

			snap := snap.NewPebble(snap.PebbleOpts{
				SnapDir: tmpDir,
			})

			_, err = snap.Revision(context.Background())
			g.Expect(err).To(HaveOccurred())
		})

		t.Run("fails when bom.json is malformed", func(t *testing.T) {
			g := NewWithT(t)

			tmpDir, err := os.MkdirTemp("", "test-bom-bad")
			g.Expect(err).To(Not(HaveOccurred()))
			defer os.RemoveAll(tmpDir)

			err = os.WriteFile(filepath.Join(tmpDir, "bom.json"), []byte("not-json"), 0o644)
			g.Expect(err).To(Not(HaveOccurred()))

			snap := snap.NewPebble(snap.PebbleOpts{
				SnapDir: tmpDir,
			})

			_, err = snap.Revision(context.Background())
			g.Expect(err).To(HaveOccurred())
		})

		t.Run("fails when revision is empty", func(t *testing.T) {
			g := NewWithT(t)

			tmpDir, err := os.MkdirTemp("", "test-bom-missing-revision")
			g.Expect(err).To(Not(HaveOccurred()))
			defer os.RemoveAll(tmpDir)

			bomContent := `{
                                "k8s": {
                                                "revision": ""
                                        }
                                }
                        }`

			err = os.WriteFile(filepath.Join(tmpDir, "bom.json"), []byte(bomContent), 0o644)
			g.Expect(err).To(Not(HaveOccurred()))

			snap := snap.NewPebble(snap.PebbleOpts{
				SnapDir: tmpDir,
			})

			_, err = snap.Revision(context.Background())
			g.Expect(err).To(HaveOccurred())
		})
	})
}
