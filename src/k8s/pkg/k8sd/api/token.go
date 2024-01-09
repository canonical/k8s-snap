package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiv1 "github.com/canonical/k8s/api/v1"
	"github.com/canonical/k8s/pkg/k8sd/api/impl"
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
	var req apiv1.CreateJoinTokenRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return response.SmartError(err)
	}

	logrus.WithField("nodeName", req.Name).Info("create token entry")
	token, err := impl.CreateJoinToken(r.Context(), s, req.Name)
	if err != nil {
		response.SmartError(fmt.Errorf("failed to create token entry: %w", err))
	}

	result := apiv1.CreateJoinTokenResponse{
		Token: token,
	}

	return response.SyncResponse(true, &result)
}
