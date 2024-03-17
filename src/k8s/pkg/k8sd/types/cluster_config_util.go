package types

import (
	"fmt"
	"slices"
)

func mergeField[T comparable](old *T, new *T, allowChange bool) (*T, error) {
	// old value is not set, use new
	if old == nil {
		return new, nil
	}
	// new value is not set, or same as old
	if new == nil || *new == *old {
		return old, nil
	}

	// both values are not-empty
	if !allowChange {
		return nil, fmt.Errorf("value has changed")
	}
	return new, nil
}

func mergeSliceField[T comparable](old *[]T, new *[]T, allowChange bool) (*[]T, error) {
	// old value is not set, use new
	if old == nil {
		return new, nil
	}
	// new value is not set, or same as old
	if new == nil || slices.Equal(*new, *old) {
		return old, nil
	}

	// both values are not-empty
	if !allowChange {
		return nil, fmt.Errorf("value has changed")
	}
	return new, nil
}

func getField[T any](val *T) T {
	if val != nil {
		return *val
	}
	var zero T
	return zero
}
