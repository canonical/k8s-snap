package cmdutil_test

import (
	"testing"

	cmdutil "github.com/canonical/k8s/cmd/util"
	. "github.com/onsi/gomega"
)

func TestEnvironWithDefaults(t *testing.T) {
	for _, tc := range []struct {
		name        string
		env         []string
		defaults    []string
		expectedEnv []string
	}{
		{
			name:        "AddMissing",
			env:         []string{"KEY1=VAL1"},
			defaults:    []string{"KEY2", "VAL2"},
			expectedEnv: []string{"KEY1=VAL1", "KEY2=VAL2"},
		},
		{
			name:        "KeepExisting",
			env:         []string{"KEY1=VAL1", "KEY2=VAL1"},
			defaults:    []string{"KEY2", "VAL2"},
			expectedEnv: []string{"KEY1=VAL1", "KEY2=VAL1"},
		},
		{
			name:        "KeepEmpty",
			env:         []string{"KEY1=VAL1", "KEY2="},
			defaults:    []string{"KEY2", "VAL2"},
			expectedEnv: []string{"KEY1=VAL1", "KEY2="},
		},
		{
			name:        "AddSome",
			env:         []string{"KEY1=VAL1", "KEY3=VAL1"},
			defaults:    []string{"KEY2", "VAL2", "KEY3", "VAL3"},
			expectedEnv: []string{"KEY1=VAL1", "KEY3=VAL1", "KEY2=VAL2"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(cmdutil.EnvironWithDefaults(tc.env, tc.defaults...)).To(Equal(tc.expectedEnv))
		})
	}
}

func TestExistsInEnviron(t *testing.T) {
	for _, tc := range []struct {
		name   string
		env    []string
		key    string
		exists bool
	}{
		{
			name:   "KeyExists",
			env:    []string{"KEY1=VAL1", "KEY2=VAL2"},
			key:    "KEY1",
			exists: true,
		},
		{
			name:   "KeyNotExists",
			env:    []string{"KEY1=VAL1", "KEY2=VAL2"},
			key:    "KEY3",
			exists: false,
		},
		{
			name:   "KeyWithEmptyValue",
			env:    []string{"KEY1=VAL1", "KEY2="},
			key:    "KEY2",
			exists: false,
		},
		{
			name:   "EmptyEnviron",
			env:    []string{},
			key:    "KEY1",
			exists: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(cmdutil.ExistsInEnviron(tc.env, tc.key)).To(Equal(tc.exists))
		})
	}
}
