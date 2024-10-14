package snaputil_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/k8s/pkg/snap/mock"
	snaputil "github.com/canonical/k8s/pkg/snap/util"
	. "github.com/onsi/gomega"
)

func TestGetServiceArgument(t *testing.T) {
	g := NewWithT(t)
	dir := t.TempDir()

	s := &mock.Snap{
		Mock: mock.Mock{
			ServiceArgumentsDir: dir,
		},
	}

	for svc, args := range map[string]string{
		"service": `
--key=value
--key-with-space value2
  --key-with-padding=value3
--multiple=keys --in-the-same-row=this-is-lost
		`,
		"service2": `
--key=value-of-service-two
`,
	} {
		g.Expect(os.WriteFile(filepath.Join(dir, svc), []byte(args), 0600)).To(Succeed())
	}

	for _, tc := range []struct {
		service     string
		key         string
		expectValue string
		expectErr   bool
	}{
		{service: "service", key: "--key", expectValue: "value"},
		{service: "service2", key: "--key", expectValue: "value-of-service-two"},
		{service: "service", key: "--key-with-padding", expectValue: "value3"},
		{service: "service", key: "--key-with-space", expectValue: "value2"},
		{service: "service", key: "--missing", expectValue: ""},
		{service: "service3", key: "--missing-service", expectValue: "", expectErr: true},
		// NOTE: the final test case documents that arguments in the same row will not be parsed properly.
		// This is carried over from the original Python code, and probably needs fixing in the future.
		{service: "service", key: "--in-the-same-row", expectValue: ""},
	} {
		t.Run(fmt.Sprintf("%s/%s", tc.service, tc.key), func(t *testing.T) {
			g := NewWithT(t)

			value, err := snaputil.GetServiceArgument(s, tc.service, tc.key)
			if tc.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(value).To(Equal(tc.expectValue))
			}
		})
	}
}

func TestUpdateServiceArguments(t *testing.T) {
	t.Run("HandleFileNotExist", func(t *testing.T) {
		g := NewWithT(t)
		s := &mock.Snap{
			Mock: mock.Mock{
				ServiceArgumentsDir: t.TempDir(),
			},
		}

		_, err := snaputil.GetServiceArgument(s, "service", "--key")
		g.Expect(err).To(HaveOccurred())

		changed, err := snaputil.UpdateServiceArguments(s, "service", map[string]string{"--key": "value"}, nil)
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(changed).To(BeTrue())

		value, err := snaputil.GetServiceArgument(s, "service", "--key")
		g.Expect(err).To(Not(HaveOccurred()))
		g.Expect(value).To(Equal("value"))
	})

	initialArguments := map[string]string{
		"--key":        "value",
		"--other":      "other-value",
		"--with-space": "value2",
	}
	for _, tc := range []struct {
		name           string
		update         map[string]string
		delete         []string
		expectedValues map[string]string
		expectedChange bool
	}{
		{
			name:   "no-change",
			update: map[string]string{"--key": "value"},
			delete: []string{"--non-existent"},
			expectedValues: map[string]string{
				"--key":   "value",
				"--other": "other-value",
			},
			expectedChange: false,
		},
		{
			name:   "no-change-space",
			update: map[string]string{"--with-space": "value2"},
			delete: []string{},
			expectedValues: map[string]string{
				"--with-space": "value2",
			},
			expectedChange: false,
		},
		{
			name:   "simple-update",
			update: map[string]string{"--key": "new-value"},
			delete: []string{},
			expectedValues: map[string]string{
				"--key":   "new-value",
				"--other": "other-value",
			},
			expectedChange: true,
		},
		{
			name:   "delete-one",
			delete: []string{"--with-space"},
			expectedValues: map[string]string{
				"--key":        "value",
				"--other":      "other-value",
				"--with-space": "",
			},
			expectedChange: true,
		},
		{
			name:   "update-many-delete-one",
			update: map[string]string{"--key": "new-value", "--other": "other-new-value"},
			delete: []string{"--with-space"},
			expectedValues: map[string]string{
				"--key":        "new-value",
				"--other":      "other-new-value",
				"--with-space": "",
			},
			expectedChange: true,
		},
		{
			name:   "update-many-single-list",
			update: map[string]string{"--key": "new-value", "--other": "other-new-value"},
			expectedValues: map[string]string{
				"--key":   "new-value",
				"--other": "other-new-value",
			},
			expectedChange: true,
		},
		{
			name: "no-updates",
			expectedValues: map[string]string{
				"--key": "value",
			},
			expectedChange: false,
		},
		{
			name:   "new-opt",
			update: map[string]string{"--new-opt": "opt-value"},
			expectedValues: map[string]string{
				"--new-opt": "opt-value",
			},
			expectedChange: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			dir := t.TempDir()

			s := &mock.Snap{
				Mock: mock.Mock{
					ServiceArgumentsDir: dir,
				},
			}
			changed, err := snaputil.UpdateServiceArguments(s, "service", initialArguments, nil)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(changed).To(BeTrue())

			changed, err = snaputil.UpdateServiceArguments(s, "service", tc.update, tc.delete)
			g.Expect(err).To(Not(HaveOccurred()))
			g.Expect(changed).To(Equal(tc.expectedChange))

			for key, expectedValue := range tc.expectedValues {
				g.Expect(snaputil.GetServiceArgument(s, "service", key)).To(Equal(expectedValue))
			}

			t.Run("Reapply", func(t *testing.T) {
				g := NewWithT(t)
				changed, err := snaputil.UpdateServiceArguments(s, "service", tc.update, tc.delete)
				g.Expect(err).To(Not(HaveOccurred()))
				g.Expect(changed).To(BeFalse())
			})
		})
	}
}
