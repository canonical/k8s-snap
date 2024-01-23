package api

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

var k8sdToken = rest.Endpoint{
	Path: "k8sd/tokens",
	Post: rest.EndpointAction{Handler: tokenPost, AllowUntrusted: true},
}

func tokenPost(s *state.State, r *http.Request) response.Response {
	return response.SmartError(fmt.Errorf("must not call"))
}
