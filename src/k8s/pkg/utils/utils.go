package utils

import "fmt"

// sanitiseMap converts a map with interface{} keys to a map with string keys.
// This is useful for preparing data for use with the Helm client, which requires
// map keys to be strings. Nested maps are also recursively processed to ensure
// all keys are converted to strings.
func SanitiseMap(m map[interface{}]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for key, value := range m {
		switch t := value.(type) {
		case map[interface{}]interface{}:
			result[fmt.Sprint(key)] = SanitiseMap(t)
		default:
			result[fmt.Sprint(key)] = value
		}
	}
	return result
}
