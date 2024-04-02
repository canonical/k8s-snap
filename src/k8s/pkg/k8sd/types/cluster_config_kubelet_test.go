package types_test

import (
	"testing"

	"github.com/canonical/k8s/pkg/k8sd/types"
	"github.com/canonical/k8s/pkg/utils/vals"
	. "github.com/onsi/gomega"
)

func TestKubelet(t *testing.T) {
	for _, tc := range []struct {
		name      string
		kubelet   types.Kubelet
		configmap map[string]string
	}{
		{
			name:      "Nil",
			configmap: map[string]string{},
		},
		{
			name: "Empty",
			configmap: map[string]string{
				"cluster-dns":    "",
				"cluster-domain": "",
				"cloud-provider": "",
			},
			kubelet: types.Kubelet{
				ClusterDNS:    vals.Pointer(""),
				ClusterDomain: vals.Pointer(""),
				CloudProvider: vals.Pointer(""),
			},
		},
		{
			name: "OnlyProvider",
			configmap: map[string]string{
				"cloud-provider": "external",
			},
			kubelet: types.Kubelet{
				CloudProvider: vals.Pointer("external"),
			},
		},
		{
			name: "OnlyDNS",
			configmap: map[string]string{
				"cluster-dns":    "1.1.1.1",
				"cluster-domain": "cluster.local",
			},
			kubelet: types.Kubelet{
				ClusterDNS:    vals.Pointer("1.1.1.1"),
				ClusterDomain: vals.Pointer("cluster.local"),
			},
		},
		{
			name: "All",
			configmap: map[string]string{
				"cluster-dns":    "1.1.1.1",
				"cluster-domain": "cluster.local",
				"cloud-provider": "external",
			},
			kubelet: types.Kubelet{
				ClusterDNS:    vals.Pointer("1.1.1.1"),
				ClusterDomain: vals.Pointer("cluster.local"),
				CloudProvider: vals.Pointer("external"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("ToConfigMap", func(t *testing.T) {
				g := NewWithT(t)

				cm, err := tc.kubelet.ToConfigMap()
				g.Expect(err).To(Succeed())
				g.Expect(cm).To(Equal(tc.configmap))
			})

			t.Run("FromConfigMap", func(t *testing.T) {
				g := NewWithT(t)

				k, err := types.KubeletFromConfigMap(tc.configmap)
				g.Expect(err).To(Succeed())
				g.Expect(k).To(Equal(tc.kubelet))
			})
		})
	}

}
