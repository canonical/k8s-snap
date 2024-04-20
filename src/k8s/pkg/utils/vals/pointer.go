package vals

func Pointer[T any](v T) *T {
	return &v
}
