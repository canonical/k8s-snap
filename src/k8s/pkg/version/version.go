package version

import (
	"encoding/json"
	"fmt"
)

const (
	// NodeAnnotationKey is the key used for the node annotation that stores the k8s snap version.
	NodeAnnotationKey = "k8sd.io/version"
)

// Info represents the version info of the k8s snap.
type Info struct {
	// Revision is the revision of the k8s snap.
	Revision string `json:"revision"`

	// NOTE(Hue): Future k8s version info can be added here.
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
