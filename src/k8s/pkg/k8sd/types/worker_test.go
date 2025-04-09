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

func TestWorkerTokenEncodeWithPadding(t *testing.T) {
	token := &types.InternalWorkerNodeToken{
		Token:         "padded-token",
		Secret:        "padded-secret",
		JoinAddresses: []string{"192.168.1.23:6400"},
		Fingerprint:   "6bf8e523089120059a4e4be99ae1250545e755d66cfa0cc32d438ea49e7575d3",
	}

	g := NewWithT(t)
	s, err := token.Encode()
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(s).ToNot(BeEmpty())

	// Ensure that the encoded string has a valid Base64 format (with padding)
	g.Expect(len(s)%4).To(Equal(0), "Base64-encoded string should have padding")

	decoded := &types.InternalWorkerNodeToken{}
	err = decoded.Decode(s)
	g.Expect(err).To(Not(HaveOccurred()))
	g.Expect(decoded).To(Equal(token))
}
