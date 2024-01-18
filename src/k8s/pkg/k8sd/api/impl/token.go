package impl

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/canonical/microcluster/state"
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

// CheckK8sdToken checks if there exists an entry in the k8sd db for this token.
func CheckK8sdToken(ctx context.Context, s *state.State, token K8sdToken) (bool, error) {
	return true, nil
	// TODO(bschimke): The k8sd token is removed after k8sd of the joining node joins the cluster.
	//                 We probably need to create an independent token entry
	/* 	isValid := false
	   	var err error
	   	logrus.WithField("token", token).Info("Check token validity")
	   	err = s.Database.Transaction(s.Context, func(ctx context.Context, tx *sql.Tx) error {
	   		records, err := cluster.GetInternalTokenRecords(ctx, tx)
	   		logrus.WithField("records", records).WithField("err", err).Info("token records")

	   		isValid, err = cluster.InternalTokenRecordExists(ctx, tx, token.Secret)
	   		return err
	   	})

	   	return isValid, err */
}
