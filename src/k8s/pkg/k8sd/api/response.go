package api

import (
	"github.com/canonical/lxd/lxd/response"
)

const (
	StatusNodeUnavailable = 520 // Node cannot be removed because it isn't in the cluster
	StatusNodeInUse       = 521 // Node cannot be joined because it is in the cluster.
)

func NodeUnavalable(err error) response.Response {
	return response.ErrorResponse(StatusNodeUnavailable, err.Error())
}

func NodeInUse(err error) response.Response {
	return response.ErrorResponse(StatusNodeInUse, err.Error())
}
