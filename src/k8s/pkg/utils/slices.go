package utils

import (
	"cmp"
	"slices"
)

// IsSubSlice checks if sub is a subslice of slice
func IsSubSlice[T cmp.Ordered](slice, sub []T) bool {
	n, m := len(slice), len(sub)

	// If the length of sub is greater than the slice, it can't be a subslice
	if m > n {
		return false
	}

	slices.Sort(slice)
	slices.Sort(sub)
	for i := 0; i <= n-m; i++ {
		if !slices.Equal(slice[i:i+m], sub) {
			return false
		}
	}

	return true
}
