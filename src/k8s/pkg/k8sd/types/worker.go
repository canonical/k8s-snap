package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// InternalWorkerNodeToken encodes information required to join a cluster as a worker node.
// InternalWorkerNodeToken encodes fields with single-letter short names to be short.
type InternalWorkerNodeToken struct {
	Token         string   `json:"token"`
	JoinAddresses []string `json:"join_addresses"`
}

// internalWorkerNodeTokenSerializeMagicString is used to validate serialized worker node tokens.
const internalWorkerNodeTokenSerializeMagicString = "m!!"

var encoding = base64.RawStdEncoding

type serializableWorkerNodeInfo struct {
	InternalWorkerNodeToken
	Magic string `json:"_"`
}

// Encode a worker node token to a base64-encoded string.
func (t *InternalWorkerNodeToken) Encode() (string, error) {
	b, err := json.Marshal(serializableWorkerNodeInfo{InternalWorkerNodeToken: *t, Magic: internalWorkerNodeTokenSerializeMagicString})
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	return encoding.EncodeToString(b), nil
}

// Decode parses the base64-encoded string into the token.
// Decode returns an error if the token is not valid.
func (t *InternalWorkerNodeToken) Decode(encoded string) error {
	raw, err := encoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("failed to deserialize token: %w", err)
	}
	var st serializableWorkerNodeInfo
	if err := json.Unmarshal(raw, &st); err != nil {
		return fmt.Errorf("failed to unmarshal token: %w", err)
	}
	if st.Magic != internalWorkerNodeTokenSerializeMagicString {
		return fmt.Errorf("magic string mismatch")
	}

	*t = st.InternalWorkerNodeToken
	return nil
}
