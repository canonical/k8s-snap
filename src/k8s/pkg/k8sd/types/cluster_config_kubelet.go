package types

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Kubelet struct {
	CloudProvider *string `json:"cloud-provider,omitempty"`
	ClusterDNS    *string `json:"cluster-dns,omitempty"`
	ClusterDomain *string `json:"cluster-domain,omitempty"`
}

func (c Kubelet) GetCloudProvider() string { return getField(c.CloudProvider) }
func (c Kubelet) GetClusterDNS() string    { return getField(c.ClusterDNS) }
func (c Kubelet) GetClusterDomain() string { return getField(c.ClusterDomain) }
func (c Kubelet) Empty() bool {
	return c.CloudProvider == nil && c.ClusterDNS == nil && c.ClusterDomain == nil
}

// hash returns a sha256 sum from the Kubelet configuration.
func (c Kubelet) hash() ([]byte, error) {
	// encoding/json.Marshal() ensures alphabetical order on JSON fields, so will
	// always produce the same JSON document.
	hash, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to hash config: %w", err)
	}

	// calculate sha256 sum of produced JSON
	h := sha256.New()
	if _, err := h.Write(hash); err != nil {
		return nil, fmt.Errorf("failed to compute sha256: %w", err)
	}
	return h.Sum(nil), nil
}

// ToConfigMap converts a Kubelet config to a map[string]string to store in a Kubernetes configmap.
// ToConfigMap will append a "k8sd-mac" field with a signed hash of the contents, if a key is specified.
func (c Kubelet) ToConfigMap(key *rsa.PrivateKey) (map[string]string, error) {
	data := make(map[string]string)

	if v := c.CloudProvider; v != nil {
		data["cloud-provider"] = *v
	}
	if v := c.ClusterDNS; v != nil {
		data["cluster-dns"] = *v
	}
	if v := c.ClusterDomain; v != nil {
		data["cluster-domain"] = *v
	}

	if key != nil {
		hash, err := c.hash()
		if err != nil {
			return nil, fmt.Errorf("failed to compute hash: %w", err)
		}
		mac, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash)
		if err != nil {
			return nil, fmt.Errorf("failed to sign hash: %w", err)
		}
		data["k8sd-mac"] = base64.StdEncoding.EncodeToString(mac)
	}

	return data, nil
}

// KubeletFromConfigMap parses configmap data into a Kubelet config.
// KubeletFromConfigMap will attempt to validate the signature (found in the "k8sd-mac" field) if a key is specified.
// KubeletFromConfigMap can parse and validate maps created with Kubelet.ToConfigMap().
func KubeletFromConfigMap(m map[string]string, key *rsa.PublicKey) (Kubelet, error) {
	var c Kubelet
	if m == nil {
		return c, nil
	}

	if v, ok := m["cloud-provider"]; ok {
		c.CloudProvider = &v
	}
	if v, ok := m["cluster-dns"]; ok {
		c.ClusterDNS = &v
	}
	if v, ok := m["cluster-domain"]; ok {
		c.ClusterDomain = &v
	}

	if key != nil {
		hash, err := c.hash()
		if err != nil {
			return Kubelet{}, fmt.Errorf("failed to compute config hash: %w", err)
		}
		signature, err := base64.StdEncoding.DecodeString(m["k8sd-mac"])
		if err != nil {
			return Kubelet{}, fmt.Errorf("failed to parse signature: %w", err)
		}
		if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, hash, signature); err != nil {
			return Kubelet{}, fmt.Errorf("failed to verify signature: %w", err)
		}
	}

	return c, nil
}
