package cmdutil_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"unicode"

	. "github.com/onsi/gomega"
)

type Main struct {
	*A
	B
}

type BoolMaybeAsString bool

func (v *BoolMaybeAsString) UnmarshalJSON(b []byte) error {
	var test bool
	if errRet := json.Unmarshal(b, &test); errRet != nil {
		if err := json.Unmarshal(bytes.Trim(b, `"`), &test); err != nil {
			return errRet
		}
	}
	*v = BoolMaybeAsString(test)
	return nil
}

type IntMaybeAsString int

func (v *IntMaybeAsString) UnmarshalJSON(b []byte) error {
	var test int
	if errRet := json.Unmarshal(b, &test); errRet != nil {
		if err := json.Unmarshal(bytes.Trim(b, `"`), &test); err != nil {
			return errRet
		}
	}
	*v = IntMaybeAsString(test)
	return nil
}

type StringSliceMaybeAsString []string

func (v *StringSliceMaybeAsString) UnmarshalJSON(b []byte) error {
	var test []string
	err := json.Unmarshal(b, &test)
	if err == nil {
		*v = test
		return nil
	}

	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return err
	}

	clean := string(bytes.Trim(b, `"`))
	if clean == "" {
		*v = nil
		return nil
	}
	*v = strings.FieldsFunc(strings.Trim(string(b), `"`), func(r rune) bool { return unicode.IsSpace(r) || r == ',' })
	return nil
}

type A struct {
	Bool *BoolMaybeAsString `json:"a.bool"`
	Str  string             `json:"a.str"`
}

type B struct {
	Bool BoolMaybeAsString        `json:"b.bool"`
	Int  IntMaybeAsString         `json:"b.int"`
	List StringSliceMaybeAsString `json:"b.list"`
}

func UnmarshalStrict(b []byte, v any) error {
	decoder := json.NewDecoder(bytes.NewBuffer(b))
	decoder.DisallowUnknownFields()

	return decoder.Decode(v)
}

func TestJSON(t *testing.T) {
	var m Main
	g := NewWithT(t)
	g.Expect(UnmarshalStrict([]byte(`{"c": true, "a.bool": ""true"", "a.str": "test", "b.bool": "true", "b.int": "10", "b.list": 13}`), &m)).To(BeNil())

	g.Expect(m).To(Equal(Main{
		A: &A{Bool: nil, Str: "test"},
		B: B{Bool: true, Int: 10},
	}))
	t.FailNow()
}
