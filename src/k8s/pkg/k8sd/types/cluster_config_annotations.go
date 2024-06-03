package types

type Annotations map[string]string

func (a Annotations) Get(key string) (value string, exists bool) {
	if a == nil {
		return "", false
	}

	v, ok := a[key]
	return v, ok
}
