package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestWorkerTokenEncode(t *testing.T) {
	token := &types.InternalWorkerNodeInfo{
		Token:         "token1",
		JoinAddresses: []string{"addr1:1010", "addr2:1212"},
	}

	g := NewWithT(t)
	s, err := token.Encode()
	g.Expect(err).To(BeNil())
	g.Expect(s).ToNot(BeEmpty())

	decoded := &types.InternalWorkerNodeInfo{}
	err = decoded.Decode(s)
	g.Expect(err).To(BeNil())

	g.Expect(decoded).To(Equal(token))
}
