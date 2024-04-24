package utils

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewStrictJSONDecoder(t *testing.T) {
	RegisterTestingT(t)

	type TestData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "Valid JSON",
			input:     `{"name": "John Doe", "age": 30}`,
			expectErr: false,
		},
		{
			name:      "JSON with unknown fields",
			input:     `{"name": "Jane Doe", "age": 25, "occupation": "Engineer"}`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			r := strings.NewReader(tt.input)
			decoder := NewStrictJSONDecoder(r)

			var data TestData
			err := decoder.Decode(&data)

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
