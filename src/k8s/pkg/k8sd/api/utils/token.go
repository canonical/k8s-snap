package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// K8sdToken is the token that is used to cluster k8sd and
// contains base64 encoded cluster information
type K8sdToken struct {
	Token         string   `json:"token"`
	NodeName      string   `json:"name"`
	Secret        string   `json:"secret"`
	Fingerprint   string   `json:"fingerprint"`
	JoinAddresses []string `json:"join_addresses"`
}

// K8sdTokenFromBase64Token creates a K8sdToken instance
// from a microcluster base64 token.
func K8sdTokenFromBase64Token(token64 string) (K8sdToken, error) {
	tokenData, err := base64.StdEncoding.DecodeString(token64)
	if err != nil {
		return K8sdToken{}, fmt.Errorf("failed to decode k8sd token %s: %w", tokenData, err)
	}

	token := K8sdToken{}
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		return K8sdToken{}, fmt.Errorf("failed to unmarshal k8sd token: %w", err)
	}
	token.Token = token64
	return token, nil
}
