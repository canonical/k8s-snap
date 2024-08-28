package utils

import (
	"cmp"
	"slices"
)

// ContainsAll checks if slice contains all elements of sub.
func ContainsAll[T cmp.Ordered](slice, sub []T) bool {
	for _, element := range sub {
		if !slices.Contains(slice, element) {
			return false
		}
	}
	return true
}
