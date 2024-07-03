package apiv1

func getField[T any](val *T) T {
	if val != nil {
		return *val
	}
	var zero T
	return zero
}
