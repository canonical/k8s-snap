package cmdutil_test

import (
	"testing"

	cmdutil "github.com/canonical/k8s/cmd/util"
	. "github.com/onsi/gomega"
)

func TestEnvWithKeyIfMissing(t *testing.T) {
	for _, tc := range []struct {
		name        string
		env         []string
		key         string
		val         string
		expectedEnv []string
	}{
		{
			name:        "AddMissing",
			env:         []string{"EXISTING=VAL"},
			key:         "NEWKEY",
			val:         "VALUE",
			expectedEnv: []string{"EXISTING=VAL", "NEWKEY=VALUE"},
		},
		{
			name:        "KeepExisting",
			env:         []string{"EXISTING=VAL", "NEWKEY=OLDVAL"},
			key:         "NEWKEY",
			val:         "VALUE",
			expectedEnv: []string{"EXISTING=VAL", "NEWKEY=OLDVAL"},
		},
		{
			name:        "KeepEmpty",
			env:         []string{"EXISTING=VAL", "NEWKEY="},
			key:         "NEWKEY",
			val:         "VALUE",
			expectedEnv: []string{"EXISTING=VAL", "NEWKEY="},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(cmdutil.EnvWithKeyIfMissing(tc.env, tc.key, tc.val)).To(Equal(tc.expectedEnv))
		})
	}
}
