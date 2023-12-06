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
	k8sdTokensEndpoint = rest.Endpoint{
		Name: "K8sdTokens",
		Path: "k8sd/tokens",

		Get: rest.EndpointAction{
			Handler: func(state *state.State, r *http.Request) response.Response {
				token := r.Header.Get("token")

				var username string
				var groups []string
				if err := state.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
					var err error
					username, groups, err = database.CheckToken(ctx, tx, token)
					return err
				}); err != nil {
					return httputil.JSONResponse(http.StatusNotFound, v1.CheckTokenResponse{Error: err.Error()})
				}

				return httputil.JSONResponse(http.StatusOK, v1.CheckTokenResponse{Username: username, Groups: groups})
			},
		},
		Post: rest.EndpointAction{
			Handler: func(state *state.State, r *http.Request) response.Response {
				request := v1.CreateTokenRequest{}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					return httputil.JSONResponse(http.StatusBadRequest, v1.CreateTokenResponse{Error: fmt.Errorf("failed to parse request: %w", err).Error()})
				}

				var token string
				if err := state.Database.Transaction(r.Context(), func(ctx context.Context, tx *sql.Tx) error {
					var err error
					token, err = database.GetOrCreateToken(ctx, tx, request.Username, request.Groups)
					return err
				}); err != nil {
					return httputil.JSONResponse(http.StatusInternalServerError, v1.CreateTokenResponse{Error: err.Error()})
				}

				return httputil.JSONResponse(http.StatusOK, v1.CreateTokenResponse{Token: token})
			},
		},
	}

	k8sdTokensWebhookEndpoint = rest.Endpoint{
		Name: "K8sdTokensWebhook",
		Path: "k8sd/tokens/webhook",

		Post: rest.EndpointAction{
			AllowUntrusted: true,
			// AccessHandler: func(state *state.State, r *http.Request) response.Response {
			// 	return response.EmptySyncResponse
			// },
			Handler: func(state *state.State, r *http.Request) response.Response {
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
			},
		},
	}
)
