package utils

import (
	"fmt"
	"testing"
	
	. "github.com/onsi/gomega"
)

func TestSnapPaths(t *testing.T) {
	t.Setenv("SNAP", "snapenv")
	t.Setenv("SNAP_DATA", "/data/snapenv")
	t.Setenv("SNAP_COMMON", "common/snapenv")
	for _, tc := range []struct {
		name          string
		input_path    []string
		expected_path string
	}{
		{
			name:          "nil",
			expected_path: "snapenv",
		},
		{
			name:          "abc",
			input_path:    []string{"abc"},
			expected_path: "snapenv/abc",
		},
		{
			name:          "abc/def",
			input_path:    []string{"abc", "def"},
			expected_path: "snapenv/abc/def",
		},
	} {
		g := NewWithT(t)
		g.Expect(SnapPath(tc.input_path...)).To(Equal(tc.expected_path))
		g.Expect(SnapDataPath(tc.input_path...)).To(Equal(fmt.Sprintf("/data/%s", tc.expected_path)))
		g.Expect(SnapCommonPath(tc.input_path...)).To(Equal(fmt.Sprintf("common/%s", tc.expected_path)))
	}
}
