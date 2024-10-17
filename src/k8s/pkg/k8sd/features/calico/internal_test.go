package calico

import (
	"testing"

	apiv1_annotations "github.com/canonical/k8s-snap-api/api/v1/annotations/calico"
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
				apiv1_annotations.AnnotationAPIServerEnabled: "true",
				apiv1_annotations.AnnotationEncapsulationV4:  "IPIP",
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
				apiv1_annotations.AnnotationEncapsulationV4: "Invalid",
			},
			expectError: true,
		},
		{
			name: "InvalidAPIServerEnabled",
			annotations: map[string]string{
				apiv1_annotations.AnnotationAPIServerEnabled: "invalid",
				apiv1_annotations.AnnotationEncapsulationV4:  "VXLAN",
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
				apiv1_annotations.AnnotationAutodetectionV4FirstFound: "true",
				apiv1_annotations.AnnotationAutodetectionV4Kubernetes: "true",
			},
			expectError: true,
		},
		{
			name: "ValidAutodetectionCidrs",
			annotations: map[string]string{
				apiv1_annotations.AnnotationAutodetectionV4CIDRs: "10.1.0.0/16,2001:0db8::/32",
			},
			expectedConfig: config{
				apiServerEnabled: false,
				encapsulationV4:  "VXLAN",
				encapsulationV6:  "VXLAN",
				autodetectionV4: map[string]any{
					"cidrs": []string{"10.1.0.0/16", "2001:0db8::/32"},
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
