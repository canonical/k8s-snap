package types

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"
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
			{name: "set-empty", new: utils.Pointer("val"), expectVal: utils.Pointer("val")},
			{name: "keep-old", old: utils.Pointer("val"), expectVal: utils.Pointer("val")},
			{name: "update", old: utils.Pointer("val"), new: utils.Pointer("newVal"), allowChange: true, expectVal: utils.Pointer("newVal")},
			{name: "update-not-allowed", old: utils.Pointer("val"), new: utils.Pointer("newVal"), expectErr: true},
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
			{name: "set-empty", new: utils.Pointer(100), expectVal: utils.Pointer(100)},
			{name: "keep-old", old: utils.Pointer(100), expectVal: utils.Pointer(100)},
			{name: "update", old: utils.Pointer(100), new: utils.Pointer(200), allowChange: true, expectVal: utils.Pointer(200)},
			{name: "update-not-allowed", old: utils.Pointer(100), new: utils.Pointer(200), expectErr: true},
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
			{name: "set-empty", new: utils.Pointer(true), expectVal: utils.Pointer(true)},
			{name: "keep-old", old: utils.Pointer(false), expectVal: utils.Pointer(false)},
			{name: "disable", old: utils.Pointer(true), new: utils.Pointer(false), allowChange: true, expectVal: utils.Pointer(false)},
			{name: "enable", old: utils.Pointer(false), new: utils.Pointer(true), allowChange: true, expectVal: utils.Pointer(true)},
			{name: "disable-not-allowed", old: utils.Pointer(true), new: utils.Pointer(false), expectErr: true},
			{name: "enable-not-allowed", old: utils.Pointer(false), new: utils.Pointer(true), expectErr: true},
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
			{name: "set-empty", new: utils.Pointer([]string{"val"}), expectVal: utils.Pointer([]string{"val"})},
			{name: "keep-old", old: utils.Pointer([]string{"val"}), expectVal: utils.Pointer([]string{"val"})},
			{name: "update", old: utils.Pointer([]string{"val"}), new: utils.Pointer([]string{"newVal"}), allowChange: true, expectVal: utils.Pointer([]string{"newVal"})},
			{name: "update-not-allowed", old: utils.Pointer([]string{"val"}), new: utils.Pointer([]string{"newVal"}), expectErr: true},
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

func Test_mergeAnnotationsField(t *testing.T) {
	for _, tc := range []struct {
		name      string
		old       Annotations
		new       Annotations
		expectErr bool
		expectVal Annotations
	}{
		{name: "keep-empty"},
		{name: "set-empty", new: Annotations{"k1": "v1"}, expectVal: Annotations{"k1": "v1"}},
		{name: "keep-old", old: Annotations{"k1": "v1"}, expectVal: Annotations{"k1": "v1"}},
		{name: "update", old: Annotations{"k1": "v1"}, new: Annotations{"k1": "v2"}, expectVal: Annotations{"k1": "v2"}},
		{name: "update-add-fields", old: Annotations{"k1": "v1"}, new: Annotations{"k1": "v2", "k2": "v2"}, expectVal: Annotations{"k1": "v2", "k2": "v2"}},
		{name: "delete-fields", old: Annotations{"k1": "v1", "k2": "v2"}, new: Annotations{"k1": "-"}, expectVal: Annotations{"k2": "v2"}},
		{name: "delete-last-field", old: Annotations{"k1": "v1"}, new: Annotations{"k1": "-"}, expectVal: Annotations{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)

			result := mergeAnnotationsField(tc.old, tc.new)
			if tc.expectVal != nil {
				g.Expect(result).To(Equal(tc.expectVal))
			} else {
				g.Expect(result).To(BeNil())
			}
		})
	}
}
