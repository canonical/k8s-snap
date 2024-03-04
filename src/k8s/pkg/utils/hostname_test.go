package utils_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/utils"
	. "github.com/onsi/gomega"
)

func TestCleanHostname(t *testing.T) {
	for _, tc := range []struct {
		hostname       string
		expectHostname string
		expectValid    bool
	}{
		{hostname: "w1", expectHostname: "w1", expectValid: true},
		{hostname: "w1.internal", expectHostname: "w1.internal", expectValid: true},
		{hostname: "w1.test.domain", expectHostname: "w1.test.domain", expectValid: true},
		{hostname: "w1-with-dash", expectHostname: "w1-with-dash", expectValid: true},
		{hostname: "Capital", expectHostname: "capital", expectValid: true},
		{hostname: "dash-end-"},
		{hostname: "dot-end."},
		{hostname: "w1-with_underscore"},
		{hostname: "spaces 123"},
		{hostname: "special!@*!^%#*&$"},
	} {
		t.Run(tc.hostname, func(t *testing.T) {
			g := NewWithT(t)
			hostname, err := utils.CleanHostname(tc.hostname)
			if tc.expectValid {
				g.Expect(err).To(BeNil())
				g.Expect(hostname).To(Equal(tc.expectHostname))
			} else {
				g.Expect(err).To(Not(BeNil()))
			}
		})
	}
}
