package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/database"
	databaseutil "github.com/canonical/k8s/pkg/k8sd/database/util"
	"github.com/canonical/k8s/pkg/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/state"
)

func (e *Endpoints) getKubernetesAuthTokens(s state.State, r *http.Request) response.Response {
	token := r.Header.Get("token")

	var username string
	var groups []string
	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		username, groups, err = database.CheckToken(ctx, tx, token)
		return err
	}); err != nil {
		return response.NotFound(err)
	}

	return response.SyncResponse(true, apiv1.CheckKubernetesAuthTokenResponse{Username: username, Groups: groups})
}

func (e *Endpoints) postKubernetesAuthTokens(s state.State, r *http.Request) response.Response {
	request := apiv1.GenerateKubernetesAuthTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	token, err := databaseutil.GetOrCreateAuthToken(r.Context(), s, request.Username, request.Groups)
	if err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, apiv1.CreateKubernetesAuthTokenResponse{Token: token})
}

func (e *Endpoints) deleteKubernetesAuthTokens(s state.State, r *http.Request) response.Response {
	request := apiv1.RevokeKubernetesAuthTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	err := databaseutil.RevokeAuthToken(r.Context(), s, request.Token)
	if err != nil {
		return response.InternalError(fmt.Errorf("failed to revoke auth token: %w", err))
	}

	return response.SyncResponse(true, nil)
}

// postKubernetesAuthWebhook is used by kube-apiserver to handle TokenReview objects.
// Note that we do not use the normal response.SyncResponse here, because it breaks the response format that kube-apiserver expects.
func (e *Endpoints) postKubernetesAuthWebhook(s state.State, r *http.Request) response.Response {
	review := apiv1.TokenReview{
		APIVersion: "authentication.k8s.io/v1",
		Kind:       "TokenReview",
	}
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		review.Status.Error = fmt.Errorf("failed to parse TokenReview: %w", err).Error()
		return utils.JSONResponse(http.StatusBadRequest, review)
	}
	// reset anything the client might be passing over in the status already
	review.Status = apiv1.TokenReviewStatus{}

	// handle APIVersion and Kind
	var apiVersionErr, kindErr error
	switch review.APIVersion {
	case "authentication.k8s.io/v1", "authentication.k8s.io/v1beta1":
	default:
		apiVersionErr = fmt.Errorf("unknown GroupVersion=%s", review.APIVersion)
		review.APIVersion = "authentication.k8s.io/v1"
	}
	switch review.Kind {
	case "TokenReview":
	default:
		kindErr = fmt.Errorf("unknown Kind=%s", review.Kind)
		review.Kind = "TokenReview"
	}
	if err := errors.Join(apiVersionErr, kindErr); err != nil {
		review.Status.Error = fmt.Errorf("invalid TokenReview: %w", err).Error()
		return utils.JSONResponse(http.StatusUnauthorized, review)
	}

	// check token
	var username string
	var groups []string
	if err := s.Database().Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		username, groups, err = database.CheckToken(ctx, tx, review.Spec.Token)
		return err
	}); err != nil {
		review.Status.Error = "invalid token"
		return utils.JSONResponse(http.StatusUnauthorized, review)
	}

	review.Status = apiv1.TokenReviewStatus{
		Audiences:     review.Spec.Audiences,
		Authenticated: true,
		User: apiv1.TokenReviewStatusUserInfo{
			UID:      username,
			Username: username,
			Groups:   groups,
		},
	}
	return utils.JSONResponse(http.StatusOK, review)
}
