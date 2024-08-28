package utils

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestIsSubSlice(t *testing.T) {
	intTests := []struct {
		name     string
		slice    []int
		sub      []int
		expected bool
	}{
		{"Subslice exists", []int{5, 4, 3, 2, 1}, []int{2, 3}, true},
		{"Subslice doesn't exist", []int{5, 4, 3, 2, 1}, []int{3, 5}, true},
		{"Subslice with different length", []int{1, 2, 3, 4, 5}, []int{1, 2, 3, 4, 5, 6}, false},
		{"Subslice not in order", []int{1, 3, 5, 2, 4}, []int{2, 3}, true},
		{"Empty slice", []int{}, []int{}, true},
		{"Empty subslice", []int{1, 2, 3, 4, 5}, []int{}, true},
		{"Subslice longer than slice", []int{1, 2}, []int{1, 2, 3}, false},
	}

	stringTests := []struct {
		name     string
		slice    []string
		sub      []string
		expected bool
	}{
		{"String subslice exists", []string{"e", "d", "c", "b", "a"}, []string{"b", "c"}, true},
		{"String subslice doesn't exist", []string{"apple", "banana", "cherry"}, []string{"banana", "date"}, false},
		{"String subslice with different length", []string{"a", "b", "c", "d"}, []string{"a", "b", "c", "d", "e"}, false},
		{"String subslice not in order", []string{"x", "z", "y"}, []string{"y", "z"}, true},
		{"Empty string slice", []string{}, []string{}, true},
		{"Empty string subslice", []string{"a", "b", "c"}, []string{}, true},
		{"String subslice longer than slice", []string{"short"}, []string{"longer", "short"}, false},
		{"String slice with duplicate elements", []string{"a", "a", "b", "b"}, []string{"a", "b"}, true},
	}

	// Test for int slices
	for _, tt := range intTests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := IsSubSlice(tt.slice, tt.sub)
			g.Expect(result).To(Equal(tt.expected), "Test failed: %s", tt.name)
		})
	}

	// Test for string slices
	for _, tt := range stringTests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := IsSubSlice(tt.slice, tt.sub)
			g.Expect(result).To(Equal(tt.expected), "Test failed: %s", tt.name)
		})
	}
}
