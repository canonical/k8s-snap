package version

import (
	"encoding/json"
	"fmt"

	versionutil "k8s.io/apimachinery/pkg/util/version"
)

const (
	// NodeAnnotationKey is the key used for the node annotation that stores the k8s snap version.
	NodeAnnotationKey = "k8sd.io/version"
)

// Info represents the version info of the k8s snap.
type Info struct {
	// Revision is the Revision of the k8s snap.
	Revision string `json:"revision,omitempty"`

	// KubernetesVersion is the version of Kubernetes included in the k8s snap.
	KubernetesVersion *versionutil.Version `json:"-"`

	// NOTE(Hue): Future k8s version info can be added here.
}

func (d Info) MarshalJSON() ([]byte, error) {
	type info Info
	aux := struct {
		info
		KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	}{info: info(d)}

	if d.KubernetesVersion != nil {
		aux.KubernetesVersion = d.KubernetesVersion.String()
	}

	return json.Marshal(aux)
}

func (d *Info) UnmarshalJSON(data []byte) error {
	type info Info
	aux := struct {
		*info
		KubernetesVersion string `json:"kubernetesVersion,omitempty"`
	}{info: (*info)(d)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.KubernetesVersion != "" {
		v, err := versionutil.Parse(aux.KubernetesVersion)
		if err != nil {
			return fmt.Errorf("failed to parse kubernetes version: %w", err)
		}
		d.KubernetesVersion = v
	}

	return nil
}

// Encode encodes the version info into a byte slice.
func (d *Info) Encode() ([]byte, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal version info: %w", err)
	}
	return b, nil
}

// Decode decodes the version info from the given byte slice.
func (d *Info) Decode(data []byte) error {
	if d == nil {
		return fmt.Errorf("version info cannot be nil, initialize it before decoding")
	}
	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}
	if err := json.Unmarshal(data, d); err != nil {
		return fmt.Errorf("failed to unmarshal version info: %w", err)
	}
	return nil
}

// String returns the string representation of the version info.
func (d Info) String() string {
	return fmt.Sprintf("Revision: %s, KubernetesVersion: %s", d.Revision, d.KubernetesVersion)
}
