package types

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Kubelet struct {
	CloudProvider      *string   `json:"cloud-provider,omitempty"`
	ClusterDNS         *string   `json:"cluster-dns,omitempty"`
	ClusterDomain      *string   `json:"cluster-domain,omitempty"`
	ControlPlaneTaints *[]string `json:"control-plane-taints,omitempty"`
}

func (c Kubelet) GetCloudProvider() string        { return getField(c.CloudProvider) }
func (c Kubelet) GetClusterDNS() string           { return getField(c.ClusterDNS) }
func (c Kubelet) GetClusterDomain() string        { return getField(c.ClusterDomain) }
func (c Kubelet) GetControlPlaneTaints() []string { return getField(c.ControlPlaneTaints) }
func (c Kubelet) Empty() bool                     { return c == Kubelet{} }

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
// ToConfigMap will append a "k8sd-mac" field if a key is specified, with a signed hash of the contents.
// ToConfigMap signes a sha256 sum of the configuration, therefore requires an ECDSA key that uses elliptic.P256() or higher.
func (c Kubelet) ToConfigMap(key *ecdsa.PrivateKey) (map[string]string, error) {
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
		if len(hash) > key.Curve.Params().BitSize/8 {
			return nil, fmt.Errorf("hash size is longer than the curve's bit-length, refusing to sign truncated hash. please specify an ECDSA key with curve P-256 or larger")
		}
		mac, err := ecdsa.SignASN1(rand.Reader, key, hash)
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
func KubeletFromConfigMap(m map[string]string, key *ecdsa.PublicKey) (Kubelet, error) {
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
		if !ecdsa.VerifyASN1(key, hash, signature) {
			return Kubelet{}, fmt.Errorf("failed to verify signature")
		}
	}

	return c, nil
}
