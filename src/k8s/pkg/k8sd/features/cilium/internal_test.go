package cilium

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestInternalConfig(t *testing.T) {
	for _, tc := range []struct {
		name           string
		annotations    map[string]string
		expectedConfig config
		expectError    bool
	}{
		{
			name:        "Empty",
			annotations: map[string]string{},
			expectedConfig: config{
				devices:             "",
				directRoutingDevice: "",
			},
			expectError: false,
		},
		{
			name: "Valid",
			annotations: map[string]string{
				annotationDevices:             "eth+ lxdbr+",
				annotationDirectRoutingDevice: "eth0",
			},
			expectedConfig: config{
				devices:             "eth+ lxdbr+",
				directRoutingDevice: "eth0",
			},
			expectError: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			annotations := make(map[string]string)
			for k, v := range tc.annotations {
				annotations[k] = v
			}

			parsed, err := internalConfig(annotations)
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(parsed).To(Equal(tc.expectedConfig))
			}
		})
	}
}
