package api

import (
	"fmt"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/rest"
	"github.com/canonical/microcluster/state"
)

// TODO(neoaggelos): remove, this is replaced with hooks
var k8sdClusterJoin = rest.Endpoint{
	Path: "k8sd/cluster/join",
	Post: rest.EndpointAction{Handler: clusterJoinPost, AllowUntrusted: false},
}

func clusterJoinPost(s *state.State, r *http.Request) response.Response {
	return response.SmartError(fmt.Errorf("must not call"))
}
