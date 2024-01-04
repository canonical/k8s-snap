package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPath(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			simple := tc.expected_path
			data := fmt.Sprintf("/data/%s", tc.expected_path)
			common := fmt.Sprintf("common/%s", tc.expected_path)
			if parsed := Path(tc.input_path...); !reflect.DeepEqual(parsed, simple) {
				t.Fatalf("expected path to be %v but it was %v instead", simple, parsed)
			}
			if parsed := DataPath(tc.input_path...); !reflect.DeepEqual(parsed, data) {
				t.Fatalf("expected data path to be %v but it was %v instead", data, parsed)
			}
			if parsed := CommonPath(tc.input_path...); !reflect.DeepEqual(parsed, common) {
				t.Fatalf("expected common path to be %v but it was %v instead", common, parsed)
			}
		})
	}
}
