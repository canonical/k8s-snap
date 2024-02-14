package client

import (
	"errors"
	"testing"

	v1 "github.com/canonical/k8s/api/v1"
	. "github.com/onsi/gomega"
)

func TestResolveError(t *testing.T) {
	g := NewWithT(t)
	myErr := errors.New("Daemon not yet initialized")
	err := resolveError(myErr)
	g.Expect(errors.Is(err, &v1.ErrNotBootstrapped{})).To(BeTrue())
}
