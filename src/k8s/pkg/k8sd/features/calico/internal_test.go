package calico

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
				apiServerEnabled: false,
				encapsulationV4:  "VXLAN",
				encapsulationV6:  "VXLAN",
			},
			expectError: false,
		},
		{
			name: "Valid",
			annotations: map[string]string{
				annotationAPIServerEnabled: "true",
				annotationEncapsulationV4:  "IPIP",
			},
			expectedConfig: config{
				apiServerEnabled: true,
				encapsulationV4:  "IPIP",
				encapsulationV6:  "VXLAN",
			},
			expectError: false,
		},
		{
			name: "InvalidEncapsulation",
			annotations: map[string]string{
				annotationEncapsulationV4: "Invalid",
			},
			expectError: true,
		},
		{
			name: "InvalidAPIServerEnabled",
			annotations: map[string]string{
				annotationAPIServerEnabled: "invalid",
				annotationEncapsulationV4:  "VXLAN",
			},
			expectedConfig: config{
				apiServerEnabled: false,
				encapsulationV4:  "VXLAN",
				encapsulationV6:  "VXLAN",
			},
			expectError: false,
		},
		{
			name: "MultipleAutodetectionV4",
			annotations: map[string]string{
				annotationAutodetectionV4Firstfound: "true",
				annotationAutodetectionV4Kubernetes: "true",
			},
			expectError: true,
		},
		{
			name: "ValidAutodetectionCidrs",
			annotations: map[string]string{
				annotationAutodetectionV4Cidrs: "10.1.0.0/16,2001:0db8::/32",
			},
			expectedConfig: config{
				apiServerEnabled: false,
				encapsulationV4:  "VXLAN",
				encapsulationV6:  "VXLAN",
				autodetectionV4: &autodetection{
					CIDRs: []string{"10.1.0.0/16", "2001:0db8::/32"},
				},
				autodetectionV6: nil,
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
