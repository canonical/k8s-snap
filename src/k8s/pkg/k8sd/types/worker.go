package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// InternalWorkerNodeInfo encodes information required to join a cluster as a worker node.
// InternalWorkerNodeInfo encodes fields with single-letter short names to be short.
type InternalWorkerNodeInfo struct {
	Token         string   `json:"token"`
	JoinAddresses []string `json:"join_addresses"`
}

// internalWorkerNodeInfoSerializeMagicString is used to validate serialized worker node tokens.
const internalWorkerNodeInfoSerializeMagicString = "m!!"

var encoding = base64.RawStdEncoding

type serializableWorkerNodeInfo struct {
	InternalWorkerNodeInfo
	Magic string `json:"_"`
}

// Encode a worker node token to a base64-encoded string.
func (t *InternalWorkerNodeInfo) Encode() (string, error) {
	b, err := json.Marshal(serializableWorkerNodeInfo{InternalWorkerNodeInfo: *t, Magic: internalWorkerNodeInfoSerializeMagicString})
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	return encoding.EncodeToString(b), nil
}

// Decode parses the base64-encoded string into the token.
// Decode returns an error if the token is not valid.
func (t *InternalWorkerNodeInfo) Decode(encoded string) error {
	raw, err := encoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("failed to deserialize token: %w", err)
	}
	var st serializableWorkerNodeInfo
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("failed to unmarshal token: %w", err)
	}
	if st.Magic != internalWorkerNodeInfoSerializeMagicString {
		return fmt.Errorf("magic string mismatch")
	}

	*t = st.InternalWorkerNodeInfo
	return nil
}
