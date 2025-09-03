package utils

import (
	. "github.com/onsi/gomega"
	"testing"
)

func TestFixInvalidIproute2JSON(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name: "Valid JSON with no changes needed",
			input: []byte(`[
				{"ifindex":4,"info_data":{"id":0,"fan-map":"example"}}
			]`),
			expected: []byte(`[
				{"ifindex":4,"info_data":{"id":0,"fan-map":"example"}}
			]`),
		},
		{
			name: "Invalid JSON with VXLAN VNI combined with fan-map",
			input: []byte(`[
				{"ifindex":4,"info_data":{"id":0fan-map,"fan-map":"example"}}
			]`),
			expected: []byte(`[
				{"ifindex":4,"info_data":{"id":"0fan-map","fan-map":"example"}}
			]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			output := fixInvalidIproute2JSON(tt.input)

			g.Expect(output).To(Equal(tt.expected))
		})
	}
}
