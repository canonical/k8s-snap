package features

func ToAnyList[T any](l []T) []any {
	out := make([]any, len(l))
	for i, v := range l {
		out[i] = v
	}
	return out
}
