package utils

import (
	"reflect"
	"testing"
)

func TestToAny(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected []any
	}{
		{[]string{"a", "b"}, []any{"a", "b"}},
		{[]int{1, 2}, []any{1, 2}},
		{[]int64{10, 20}, []any{int64(10), int64(20)}},
		{[]float64{1.1, 2.2}, []any{1.1, 2.2}},
		{[]bool{true, false}, []any{true, false}},
	}

	for _, tt := range tests {
		var got []any
		switch v := tt.input.(type) {
		case []string:
			got = ToAny(v)
		case []int:
			got = ToAny(v)
		case []int64:
			got = ToAny(v)
		case []float64:
			got = ToAny(v)
		case []bool:
			got = ToAny(v)
		}
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("ToAny(%v) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}

func TestEnsureAnySlice(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected []any
	}{
		{[]any{1, 2}, []any{1, 2}},
		{[]string{"a"}, []any{"a"}},
		{[]int{3, 4}, []any{3, 4}},
		{[]int64{10}, []any{int64(10)}},
		{[]float64{1.1}, []any{1.1}},
		{[]bool{true}, []any{true}},
		{[]map[string]any{{"foo": 1}}, []any{map[string]any{"foo": 1}}},
		{[]map[string]interface{}{{"bar": 2}}, []any{map[string]interface{}{"bar": 2}}},
		{123, []any{123}},
		{nil, nil},
	}

	for _, tt := range tests {
		got := EnsureAnySlice(tt.input)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("EnsureAnySlice(%v) = %v; want %v", tt.input, got, tt.expected)
		}
	}
}
