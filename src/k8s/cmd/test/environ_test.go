package test_test

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
