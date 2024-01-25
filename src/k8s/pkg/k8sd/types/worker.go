package types

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// WorkerNodeToken encodes information required to join a cluster as a worker node.
// WorkerNodeToken encodes fields with single-letter short names to be short.
type WorkerNodeToken struct {
	CA             string   `json:"a,omitempty"`
	APIServers     []string `json:"b"`
	KubeletToken   string   `json:"c"`
	KubeProxyToken string   `json:"d"`
	ClusterCIDR    string   `json:"e"`
	ClusterDNS     string   `json:"f,omitempty"`
	ClusterDomain  string   `json:"g,omitempty"`
	CloudProvider  string   `json:"h,omitempty"`
}

// workerNodeSerializeMagicString is used to validate serialized worker node tokens.
const workerNodeSerializeMagicString = "M!"

var encoding = base64.RawStdEncoding

type serializableWorkerNodeToken struct {
	WorkerNodeToken
	Magic string `json:"_"`
}

// Encode a worker node token to a base64-encoded string.
func (t *WorkerNodeToken) Encode() (string, error) {
	var buf bytes.Buffer
	b, err := json.Marshal(serializableWorkerNodeToken{WorkerNodeToken: *t, Magic: workerNodeSerializeMagicString})
	if err != nil {
		return "", fmt.Errorf("failed to marshal token: %w", err)
	}
	encoder, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", fmt.Errorf("failed to initialize encoder: %w", err)
	}
	if _, err := encoder.Write(b); err != nil {
		return "", fmt.Errorf("failed to serialize token: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return "", fmt.Errorf("failed to serialize token: %w", err)
	}
	return encoding.EncodeToString(buf.Bytes()), nil
}

// Decode parses the base64-encoded string into the token.
// Decode returns an error if the token is not valid.
func (t *WorkerNodeToken) Decode(encoded string) error {
	raw, err := encoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("failed to deserialize token: %w", err)
	}
	decoder, err := gzip.NewReader(bytes.NewBuffer(raw))
	if err != nil {
		return fmt.Errorf("failed to initialize decoder: %w", err)
	}
	b, err := io.ReadAll(decoder)
	if err != nil {
		return fmt.Errorf("failed to deserialize token: %w", err)
	}

	var st serializableWorkerNodeToken
	if err := json.Unmarshal([]byte(b), &st); err != nil {
		return fmt.Errorf("failed to unmarshal token: %w", err)
	}
	if st.Magic != workerNodeSerializeMagicString {
		return fmt.Errorf("worker node token magic string mismatch")
	}

	*t = st.WorkerNodeToken
	return nil
}
