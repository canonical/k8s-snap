package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/httputil"
	"github.com/canonical/k8s/pkg/k8sd/database"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var (
	kubernetesAuthTokens = rest.Endpoint{
		Name: "KubernetesAuthTokens",
		Path: "kubernetes/auth/tokens",
		Get:  rest.EndpointAction{Handler: getKubernetesAuthToken, AllowUntrusted: true},
		Post: rest.EndpointAction{Handler: postKubernetesAuthToken},
	}
	kubernetesAuthWebhook = rest.Endpoint{
		Name: "KubernetesAuthWebhook",
		Path: "kubernetes/auth/webhook",
		Post: rest.EndpointAction{Handler: kubernetesAuthTokenReviewWebhook, AllowUntrusted: true},
	}
)

func getKubernetesAuthToken(state *state.State, r *http.Request) response.Response {
	token := r.Header.Get("token")

	var username string
	var groups []string
	if err := state.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		username, groups, err = database.CheckToken(ctx, tx, token)
		return err
	}); err != nil {
		return response.NotFound(err)
	}

	return response.SyncResponse(true, v1.CheckKubernetesAuthTokenResponse{Username: username, Groups: groups})
}

func postKubernetesAuthToken(state *state.State, r *http.Request) response.Response {
	request := v1.CreateKubernetesAuthTokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return response.BadRequest(fmt.Errorf("failed to parse request: %w", err))
	}

	var token string
	if err := state.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		token, err = database.GetOrCreateToken(ctx, tx, request.Username, request.Groups)
		return err
	}); err != nil {
		return response.InternalError(err)
	}

	return response.SyncResponse(true, v1.CreateKubernetesAuthTokenResponse{Token: token})
}

// kubernetesAuthTokenReviewWebhook is used by kube-apiserver to handle TokenReview objects.
// Note that we do not use the normal response.SyncResponse here, because it breaks the response format that kube-apiserver expects.
func kubernetesAuthTokenReviewWebhook(state *state.State, r *http.Request) response.Response {
	review := v1.TokenReview{
		APIVersion: "authentication.k8s.io/v1",
		Kind:       "TokenReview",
	}
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		review.Status.Error = fmt.Errorf("failed to parse TokenReview: %w", err).Error()
		return httputil.JSONResponse(http.StatusBadRequest, review)
	}
	// reset anything the client might be passing over in the status already
	review.Status = v1.TokenReviewStatus{}

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
		return httputil.JSONResponse(http.StatusUnauthorized, review)
	}

	// check token
	var username string
	var groups []string
	if err := state.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
		var err error
		username, groups, err = database.CheckToken(ctx, tx, review.Spec.Token)
		return err
	}); err != nil {
		review.Status.Error = "invalid token"
		return httputil.JSONResponse(http.StatusUnauthorized, review)
	}

	review.Status = v1.TokenReviewStatus{
		Audiences:     review.Spec.Audiences,
		Authenticated: true,
		User: v1.TokenReviewStatusUserInfo{
			UID:      username,
			Username: username,
			Groups:   groups,
		},
	}
	return httputil.JSONResponse(http.StatusOK, review)
}
