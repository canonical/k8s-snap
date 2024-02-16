package types

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_mergeValue(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			old         string
			new         string
			allowChange bool
			expectErr   bool
			expectVal   string
		}{
			{name: "set-empty", new: "val", expectVal: "val"},
			{name: "keep-old", old: "val", expectVal: "val"},
			{name: "update", old: "val", new: "newVal", allowChange: true, expectVal: "newVal"},
			{name: "update-not-allowed", old: "val", new: "newVal", expectErr: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				result, err := mergeValue(tc.old, tc.new, tc.allowChange)
				if tc.expectErr {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
					g.Expect(result).To(Equal(tc.expectVal))
				}
			})
		}
	})

	t.Run("int", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			old         int
			new         int
			allowChange bool
			expectErr   bool
			expectVal   int
		}{
			{name: "set-empty", new: 100, expectVal: 100},
			{name: "keep-old", old: 100, expectVal: 100},
			{name: "update", old: 100, new: 200, allowChange: true, expectVal: 200},
			{name: "update-not-allowed", old: 100, new: 200, expectErr: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				result, err := mergeValue(tc.old, tc.new, tc.allowChange)
				if tc.expectErr {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
					g.Expect(result).To(Equal(tc.expectVal))
				}
			})
		}
	})
}
