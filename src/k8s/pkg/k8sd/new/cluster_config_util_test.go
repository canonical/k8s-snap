package newtypes

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func Test_mergeField(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			old         *string
			new         *string
			allowChange bool
			expectErr   bool
			expectVal   *string
		}{
			{name: "keep-empty"},
			{name: "set-empty", new: vals.Pointer("val"), expectVal: vals.Pointer("val")},
			{name: "keep-old", old: vals.Pointer("val"), expectVal: vals.Pointer("val")},
			{name: "update", old: vals.Pointer("val"), new: vals.Pointer("newVal"), allowChange: true, expectVal: vals.Pointer("newVal")},
			{name: "update-not-allowed", old: vals.Pointer("val"), new: vals.Pointer("newVal"), expectErr: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				result, err := mergeField(tc.old, tc.new, tc.allowChange)
				switch {
				case tc.expectErr:
					g.Expect(err).ToNot(BeNil())
				case tc.expectVal == nil:
					g.Expect(err).To(BeNil())
					g.Expect(result).To(BeNil())
				case tc.expectVal != nil:
					g.Expect(err).To(BeNil())
					g.Expect(*result).To(Equal(*tc.expectVal))
				}
			})
		}
	})

	t.Run("int", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			old         *int
			new         *int
			allowChange bool
			expectErr   bool
			expectVal   *int
		}{
			{name: "keep-empty"},
			{name: "set-empty", new: vals.Pointer(100), expectVal: vals.Pointer(100)},
			{name: "keep-old", old: vals.Pointer(100), expectVal: vals.Pointer(100)},
			{name: "update", old: vals.Pointer(100), new: vals.Pointer(200), allowChange: true, expectVal: vals.Pointer(200)},
			{name: "update-not-allowed", old: vals.Pointer(100), new: vals.Pointer(200), expectErr: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				result, err := mergeField(tc.old, tc.new, tc.allowChange)
				switch {
				case tc.expectErr:
					g.Expect(err).ToNot(BeNil())
				case tc.expectVal == nil:
					g.Expect(err).To(BeNil())
					g.Expect(result).To(BeNil())
				case tc.expectVal != nil:
					g.Expect(err).To(BeNil())
					g.Expect(*result).To(Equal(*tc.expectVal))
				}
			})
		}
	})

	t.Run("bool", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			old         *bool
			new         *bool
			allowChange bool
			expectErr   bool
			expectVal   *bool
		}{
			{name: "keep-empty"},
			{name: "set-empty", new: vals.Pointer(true), expectVal: vals.Pointer(true)},
			{name: "keep-old", old: vals.Pointer(false), expectVal: vals.Pointer(false)},
			{name: "disable", old: vals.Pointer(true), new: vals.Pointer(false), allowChange: true, expectVal: vals.Pointer(false)},
			{name: "enable", old: vals.Pointer(false), new: vals.Pointer(true), allowChange: true, expectVal: vals.Pointer(true)},
			{name: "disable-not-allowed", old: vals.Pointer(true), new: vals.Pointer(false), expectErr: true},
			{name: "enable-not-allowed", old: vals.Pointer(false), new: vals.Pointer(true), expectErr: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				result, err := mergeField(tc.old, tc.new, tc.allowChange)
				switch {
				case tc.expectErr:
					g.Expect(err).ToNot(BeNil())
				case tc.expectVal == nil:
					g.Expect(err).To(BeNil())
					g.Expect(result).To(BeNil())
				case tc.expectVal != nil:
					g.Expect(err).To(BeNil())
					g.Expect(*result).To(Equal(*tc.expectVal))
				}
			})
		}
	})
}

func Test_mergeSliceField(t *testing.T) {
	t.Run("[]string", func(t *testing.T) {
		for _, tc := range []struct {
			name        string
			old         *[]string
			new         *[]string
			allowChange bool
			expectErr   bool
			expectVal   *[]string
		}{
			{name: "keep-empty"},
			{name: "set-empty", new: vals.Pointer([]string{"val"}), expectVal: vals.Pointer([]string{"val"})},
			{name: "keep-old", old: vals.Pointer([]string{"val"}), expectVal: vals.Pointer([]string{"val"})},
			{name: "update", old: vals.Pointer([]string{"val"}), new: vals.Pointer([]string{"newVal"}), allowChange: true, expectVal: vals.Pointer([]string{"newVal"})},
			{name: "update-not-allowed", old: vals.Pointer([]string{"val"}), new: vals.Pointer([]string{"newVal"}), expectErr: true},
		} {
			t.Run(tc.name, func(t *testing.T) {
				g := NewWithT(t)
				result, err := mergeSliceField(tc.old, tc.new, tc.allowChange)
				switch {
				case tc.expectErr:
					g.Expect(err).ToNot(BeNil())
				case tc.expectVal == nil:
					g.Expect(err).To(BeNil())
					g.Expect(result).To(BeNil())
				case tc.expectVal != nil:
					g.Expect(err).To(BeNil())
					g.Expect(*result).To(Equal(*tc.expectVal))
				}
			})
		}
	})
}
