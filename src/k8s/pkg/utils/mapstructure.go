package utils

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

// YAMLToStringSliceHookFunc returns a mapstructure.DecodeHookFunc that converts string to []string by parsing YAML.
func YAMLToStringSliceHookFunc(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Slice {
		return data, nil
	}

	if data.(string) == "" {
		return data, nil
	}

	var result []string
	if err := yaml.Unmarshal([]byte(data.(string)), &result); err != nil {
		return data, nil
	}

	return result, nil
}

// StringToFieldsSliceHookFunc is like mapstructure.StringToSliceHookFunc() but uses strings.Fields() and filters whitespace.
func StringToFieldsSliceHookFunc(r rune) mapstructure.DecodeHookFunc {
	return func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.FieldsFunc(raw, func(this rune) bool { return this == r || unicode.IsSpace(this) }), nil
	}
}

// YAMLToStringMapHookFunc returns a mapstructure.DecodeHookFunc that converts string to map[string]string by parsing YAML.
func YAMLToStringMapHookFunc(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Map {
		return data, nil
	}

	if data.(string) == "" {
		return map[string]string{}, nil
	}

	var result map[string]string
	if err := yaml.Unmarshal([]byte(data.(string)), &result); err != nil {
		return data, nil
	}

	return result, nil
}

// StringToStringMapHookFunc is like StringToFieldsSliceHookFunc(). It splits map entries on ',' and then key-value pairs on '='.
func StringToStringMapHookFunc(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	if f != reflect.String || t != reflect.Map {
		return data, nil
	}

	raw := data.(string)
	if raw == "" {
		return map[string]string{}, nil
	}

	fields := strings.FieldsFunc(raw, func(this rune) bool { return this == ',' || unicode.IsSpace(this) })
	result := make(map[string]string, len(fields))
	for _, kv := range strings.FieldsFunc(raw, func(this rune) bool { return this == ',' || unicode.IsSpace(this) }) {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) < 2 {
			return data, nil
		}
		result[parts[0]] = parts[1]
	}

	return result, nil
}
