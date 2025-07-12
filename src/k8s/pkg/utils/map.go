package utils

// DeepCopyMap copies a string map by creating a distinct new map and copying over the values.
func DeepCopyMap(original map[string]string) map[string]string {
	copied := make(map[string]string, len(original))
	for k, v := range original {
		copied[k] = v
	}
	return copied
}
