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
		return nil, fmt.Errorf("value cannot change")
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

func mergeAnnotationsField(old Annotations, new Annotations) Annotations {
	// old value is not set, use new
	if old == nil {
		return new
	}
	// new value is not set, use old
	if new == nil {
		return old
	}

	// merge fields, start from old and then add new
	// if any field is set to "-", delete it from the final result
	m := make(map[string]string, len(old)+len(new))
	for k, v := range old {
		m[k] = v
	}
	for k, v := range new {
		if v == "-" {
			delete(m, k)
		} else {
			m[k] = v
		}
	}

	return Annotations(m)
}

func getField[T any](val *T) T {
	if val != nil {
		return *val
	}
	var zero T
	return zero
}
