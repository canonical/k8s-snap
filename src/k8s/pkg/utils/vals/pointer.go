package vals

func Pointer[T any](v T) *T {
	return &v
}

func OptionalBool(v *bool, defaultValue bool) bool {
	if v != nil {
		return *v
	}
	return defaultValue
}
