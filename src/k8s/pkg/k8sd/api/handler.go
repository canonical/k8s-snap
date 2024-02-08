package api

import (
	"net/http"

	"github.com/canonical/lxd/lxd/response"
	"github.com/canonical/microcluster/microcluster"
	"github.com/canonical/microcluster/state"
)

// handler is the handler type for microcluster endpoints.
type handler func(*state.State, *http.Request) response.Response

// handlerWithMicroCluster is the handler type for endpoints that also need access to the microcluster instance.
type handlerWithMicroCluster func(*microcluster.MicroCluster, *state.State, *http.Request) response.Response

// wrapHandlerWithMicroCluster creates a microcluster handler from a handlerWithMicroCluster by capturing the microcluster instance.
func wrapHandlerWithMicroCluster(m *microcluster.MicroCluster, handler handlerWithMicroCluster) handler {
	return func(s *state.State, r *http.Request) response.Response {
		return handler(m, s, r)
	}
}
