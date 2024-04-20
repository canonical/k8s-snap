package types_test

import (
	"crypto/rand"
	"crypto/rsa"
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

				cm, err := tc.kubelet.ToConfigMap(nil)
				g.Expect(err).To(BeNil())
				g.Expect(cm).To(Equal(tc.configmap))
			})

			t.Run("FromConfigMap", func(t *testing.T) {
				g := NewWithT(t)

				k, err := types.KubeletFromConfigMap(tc.configmap, nil)
				g.Expect(err).To(BeNil())
				g.Expect(k).To(Equal(tc.kubelet))
			})
		})
	}
}

func TestKubeletSign(t *testing.T) {
	g := NewWithT(t)
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	g.Expect(err).To(BeNil())

	kubelet := types.Kubelet{
		CloudProvider: vals.Pointer("external"),
		ClusterDNS:    vals.Pointer("10.0.0.1"),
		ClusterDomain: vals.Pointer("cluster.local"),
	}

	configmap, err := kubelet.ToConfigMap(key)
	g.Expect(err).To(BeNil())
	g.Expect(configmap).To(HaveKeyWithValue("k8sd-mac", Not(BeEmpty())))

	t.Run("NoSign", func(t *testing.T) {
		g := NewWithT(t)

		configmap, err := kubelet.ToConfigMap(nil)
		g.Expect(err).To(BeNil())
		g.Expect(configmap).To(Not(HaveKey("k8sd-mac")))
	})

	t.Run("SignAndVerify", func(t *testing.T) {
		g := NewWithT(t)

		fromKubelet, err := types.KubeletFromConfigMap(configmap, &key.PublicKey)
		g.Expect(err).To(BeNil())
		g.Expect(fromKubelet).To(Equal(kubelet))
	})

	t.Run("DeterministicSignature", func(t *testing.T) {
		g := NewWithT(t)

		configmap2, err := kubelet.ToConfigMap(key)
		g.Expect(err).To(BeNil())
		g.Expect(configmap2).To(Equal(configmap))
	})

	t.Run("WrongKey", func(t *testing.T) {
		g := NewWithT(t)

		wrongKey, err := rsa.GenerateKey(rand.Reader, 2048)
		g.Expect(err).To(BeNil())

		cm, err := types.KubeletFromConfigMap(configmap, &wrongKey.PublicKey)
		g.Expect(cm).To(BeZero())
		g.Expect(err).To(HaveOccurred())
	})

	t.Run("BadSignature", func(t *testing.T) {
		for editKey := range configmap {
			t.Run(editKey, func(t *testing.T) {
				g := NewWithT(t)
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				g.Expect(err).To(BeNil())

				c, err := kubelet.ToConfigMap(key)
				g.Expect(err).To(BeNil())
				g.Expect(c).To(HaveKeyWithValue("k8sd-mac", Not(BeEmpty())))

				t.Run("Manipulated", func(t *testing.T) {
					g := NewWithT(t)
					c[editKey] = "attack"

					k, err := types.KubeletFromConfigMap(c, &key.PublicKey)
					g.Expect(err).To(HaveOccurred())
					g.Expect(k).To(BeZero())
				})

				t.Run("Deleted", func(t *testing.T) {
					g := NewWithT(t)
					delete(c, editKey)

					k, err := types.KubeletFromConfigMap(c, &key.PublicKey)
					g.Expect(err).To(HaveOccurred())
					g.Expect(k).To(BeZero())
				})
			})
		}
	})
}
