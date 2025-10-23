package utils

// ToAny converts a slice of any comparable type to a []any.
// Works for []int, []float64, []string, []bool, etc.
func ToAny[T any](s []T) []any {
	if s == nil {
		return nil
	}
	a := make([]any, len(s))
	for i, v := range s {
		a[i] = v
	}
	return a
}

// ToAnyMapsAny converts a slice of map[string]any into a []any.
func ToAnyMapsAny(m []map[string]any) []any {
	if m == nil {
		return nil
	}
	a := make([]any, len(m))
	for i, v := range m {
		a[i] = v
	}
	return a
}

// EnsureAnySlice converts an arbitrary value into []any.
// Recognized slices: []any, []string, []int, []float64, []bool, []map[string]any, []map[string]interface{}.
// Single values are wrapped into a single-element slice.
func EnsureAnySlice(v any) []any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []any:
		return t
	case []string:
		return ToAny(t)
	case []int:
		return ToAny(t)
	case []int64:
		return ToAny(t)
	case []float64:
		return ToAny(t)
	case []bool:
		return ToAny(t)
	case []map[string]any:
		return ToAnyMapsAny(t)
	default:
		return []any{t}
	}
}
