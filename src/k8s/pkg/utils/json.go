package utils

import (
	"encoding/json"
	"io"
)

// NewStrictJSONDecoder creates a new JSON decoder that disallows unknown fields.
func NewStrictJSONDecoder(r io.Reader) *json.Decoder {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	return decoder
}
