package api

import (
	"encoding/json"
	"net/http"

	api "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/utils"
	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
	"github.com/sirupsen/logrus"
)

var k8sdToken = rest.Endpoint{
	Path: "k8sd/tokens",
	Post: rest.EndpointAction{Handler: tokenPost, AllowUntrusted: true},
}

func tokenPost(s *state.State, r *http.Request) response.Response {
	// Decode the POST body to get the node name.
	var req api.CreateJoinTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(err)
	}

	logrus.WithField("nodeName", req.Name).Info("create token entry")
	token, err := utils.CreateJoinToken(r.Context(), s, req.Name)
	if err != nil {
		response.SmartError(err)
	}

	result := api.CreateJoinTokenResponse{
		Token: token,
	}

	return response.SyncResponse(true, &result)
}
