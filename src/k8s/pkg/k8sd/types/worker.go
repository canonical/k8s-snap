package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// InternalWorkerNodeToken encodes information required to join a cluster as a worker node.
type InternalWorkerNodeToken struct {
	// Secret is used to verify the join request.
	// Secret is only valid for a specified worker that was specified upon creation.
	Secret string `json:"secret"`
	// JoinAddresses is a list of control-plane addresses that exist in the cluster.
	JoinAddresses []string `json:"join_addresses"`
}

var encoding = base64.RawStdEncoding

// Encode a worker node token to a base64-encoded string.
func (t *InternalWorkerNodeToken) Encode() (string, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	return encoding.EncodeToString(b), nil
}

// Decode parses the base64-encoded string into the token.
func (t *InternalWorkerNodeToken) Decode(encoded string) error {
	raw, err := encoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("failed to deserialize token: %w", err)
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return nil
}
