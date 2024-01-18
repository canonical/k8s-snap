package snap_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/canonical/k8s/pkg/snap"
	"github.com/canonical/k8s/pkg/snap/mock"

	. "github.com/onsi/gomega"
)

func TestGetServiceArgument(t *testing.T) {
	serviceOneArguments := `
--key=value
--key-with-space value2
   --key-with-padding=value3
--multiple=keys --in-the-same-row=this-is-lost
`
	serviceTwoArguments := `
--key=value-of-service-two
`
	s := &mock.Snap{
		ServiceArguments: map[string]string{
			"service":  serviceOneArguments,
			"service2": serviceTwoArguments,
		},
	}
	if err := os.MkdirAll("testdata/args", 0755); err != nil {
		t.Fatal("Failed to setup test directory")
	}
	for _, tc := range []struct {
		service       string
		key           string
		expectedValue string
	}{
		{service: "service", key: "--key", expectedValue: "value"},
		{service: "service2", key: "--key", expectedValue: "value-of-service-two"},
		{service: "service", key: "--key-with-padding", expectedValue: "value3"},
		{service: "service", key: "--key-with-space", expectedValue: "value2"},
		{service: "service", key: "--missing", expectedValue: ""},
		{service: "service3", key: "--missing-service", expectedValue: ""},
		// NOTE: the final test case documents that arguments in the same row will not be parsed properly.
		// This is carried over from the original Python code, and probably needs fixing in the future.
		{service: "service", key: "--in-the-same-row", expectedValue: ""},
	} {
		t.Run(fmt.Sprintf("%s/%s", tc.service, tc.key), func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(snap.GetServiceArgument(s, tc.service, tc.key)).To(Equal(tc.expectedValue))
		})
	}
}

type mockSnapFileNotExist struct {
	mock.Snap
}

func (s *mockSnapFileNotExist) ReadServiceArguments(serviceName string) (string, error) {
	_, err := os.ReadFile("testdata/fileThatDoesNotExist")
	return "", fmt.Errorf("wrapped not found error: %w", err)
}

func TestUpdateServiceArguments(t *testing.T) {
	t.Run("HandleFileNotExist", func(t *testing.T) {
		g := NewWithT(t)
		s := &mockSnapFileNotExist{
			Snap: mock.Snap{},
		}

		changed, err := snap.UpdateServiceArguments(s, "service", []map[string]string{{"--key": "value"}}, nil)
		g.Expect(err).To(BeNil())
		g.Expect(changed).To(BeTrue())

		g.Expect(s.Snap.ServiceArguments["service"]).To(Equal("--key=value\n"))
	})

	initialArguments := `
--key=value
--other=other-value
--with-space value2
`
	for _, tc := range []struct {
		name           string
		update         []map[string]string
		delete         []string
		expectedValues map[string]string
		expectedChange bool
	}{
		{
			name:   "no-change",
			update: []map[string]string{{"--key": "value"}},
			delete: []string{"--non-existent"},
			expectedValues: map[string]string{
				"--key":   "value",
				"--other": "other-value",
			},
			expectedChange: false,
		},
		{
			name:   "no-change-space",
			update: []map[string]string{{"--with-space": "value2"}},
			delete: []string{},
			expectedValues: map[string]string{
				"--with-space": "value2",
			},
			expectedChange: false,
		},
		{
			name:   "simple-update",
			update: []map[string]string{{"--key": "new-value"}},
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
			update: []map[string]string{{"--key": "new-value"}, {"--other": "other-new-value"}},
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
			update: []map[string]string{{"--key": "new-value", "--other": "other-new-value"}},
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
			update: []map[string]string{{"--new-opt": "opt-value"}},
			expectedValues: map[string]string{
				"--new-opt": "opt-value",
			},
			expectedChange: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			s := &mock.Snap{
				ServiceArguments: map[string]string{
					"service": initialArguments,
				},
			}

			changed, err := snap.UpdateServiceArguments(s, "service", tc.update, tc.delete)
			g.Expect(err).To(BeNil())
			g.Expect(changed).To(Equal(tc.expectedChange))

			for key, expectedValue := range tc.expectedValues {
				g.Expect(snap.GetServiceArgument(s, "service", key)).To(Equal(expectedValue))
			}

			t.Run("Reapply", func(t *testing.T) {
				g := NewWithT(t)
				changed, err := snap.UpdateServiceArguments(s, "service", tc.update, tc.delete)
				g.Expect(err).To(BeNil())
				g.Expect(changed).To(BeFalse())
			})
		})
	}
}
