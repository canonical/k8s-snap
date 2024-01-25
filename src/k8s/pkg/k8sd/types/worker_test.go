package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	. "github.com/onsi/gomega"
)

func TestWorkerTokenEncode(t *testing.T) {
	token := &types.WorkerNodeToken{
		CA:             "CA DATA",
		APIServers:     []string{"1.1.1.1:6443"},
		ClusterCIDR:    "10.1.0.0/16",
		KubeletToken:   "token1",
		KubeProxyToken: "token2",
		ClusterDomain:  "cluster.local",
		ClusterDNS:     "10.152.183.10/24",
		CloudProvider:  "external",
	}

	g := NewWithT(t)
	s, err := token.Encode()
	g.Expect(err).To(BeNil())
	g.Expect(s).ToNot(BeEmpty())

	decoded := &types.WorkerNodeToken{}
	err = decoded.Decode(s)
	g.Expect(err).To(BeNil())

	g.Expect(decoded).To(Equal(token))
}
