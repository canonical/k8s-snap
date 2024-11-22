package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestWorkerTokenEncode(t *testing.T) {
	token := &types.InternalWorkerNodeToken{
		Token:         "token1",
		Secret:        "mysecret",
		JoinAddresses: []string{"addr1:1010", "addr2:1212"},
		Fingerprint:   "fingerprint",
	}

	g := NewWithT(t)
	s, err := token.Encode()
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(s).ToNot(BeEmpty())

	decoded := &types.InternalWorkerNodeToken{}
	err = decoded.Decode(s)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(decoded).To(Equal(token))
}
