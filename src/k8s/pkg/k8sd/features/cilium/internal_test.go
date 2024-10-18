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
			name:           "Empty annotations",
			annotations:    map[string]string{},
			expectedConfig: config{},
			expectError:    false,
		},
		{
			name: "Single valid VLAN",
			annotations: map[string]string{
				annotationVLANBPFBypass: "1",
			},
			expectedConfig: config{
				vlanBPFBypass: []int{1},
			},
			expectError: false,
		},
		{
			name: "Multiple valid VLANs",
			annotations: map[string]string{
				annotationVLANBPFBypass: "1,2,3,4,5",
			},
			expectedConfig: config{
				vlanBPFBypass: []int{1, 2, 3, 4, 5},
			},
			expectError: false,
		},
		{
			name: "Wildcard VLAN",
			annotations: map[string]string{
				annotationVLANBPFBypass: "0",
			},
			expectedConfig: config{
				vlanBPFBypass: []int{0},
			},
			expectError: false,
		},
		{
			name: "Too many VLANs",
			annotations: map[string]string{
				annotationVLANBPFBypass: "1,2,3,4,5,6",
			},
			expectError: true,
		},
		{
			name: "Invalid VLAN tag format",
			annotations: map[string]string{
				annotationVLANBPFBypass: "abc",
			},
			expectError: true,
		},
		{
			name: "VLAN tag out of range",
			annotations: map[string]string{
				annotationVLANBPFBypass: "4095",
			},
			expectError: true,
		},
		{
			name: "VLAN tag negative",
			annotations: map[string]string{
				annotationVLANBPFBypass: "-1",
			},
			expectError: true,
		},
		{
			name: "Duplicate VLAN tags",
			annotations: map[string]string{
				annotationVLANBPFBypass: "1,2,2,3",
			},
			expectError: true,
		},
		{
			name: "Mixed spaces and commas",
			annotations: map[string]string{
				annotationVLANBPFBypass: " 1, 2,3 ,4 , 5 ",
			},
			expectedConfig: config{
				vlanBPFBypass: []int{1, 2, 3, 4, 5},
			},
			expectError: false,
		},
		{
			name: "Invalid mixed with valid",
			annotations: map[string]string{
				annotationVLANBPFBypass: "1,abc,3",
			},
			expectError: true,
		},
		{
			name:           "Nil annotations",
			annotations:    nil,
			expectedConfig: config{},
			expectError:    false,
		},
		{
			name: "With curly braces",
			annotations: map[string]string{
				annotationVLANBPFBypass: "{1,2,3}",
			},
			expectedConfig: config{
				vlanBPFBypass: []int{1, 2, 3},
			},
			expectError: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			parsed, err := internalConfig(tc.annotations)
			if tc.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(parsed).To(Equal(tc.expectedConfig))
			}
		})
	}
}
