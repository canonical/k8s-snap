package api

import (
	"github.com/canonical/lxd/lxd/response"
)

const (
	// StatusNodeUnavailable is the Http status code that the API returns if the node isn't in the cluster.
	StatusNodeUnavailable = 520
	// StatusNodeInUse is the Http status code that the API returns if the node is already in the cluster.
	StatusNodeInUse = 521
)

func NodeUnavailable(err error) response.Response {
	return response.ErrorResponse(StatusNodeUnavailable, err.Error())
}

func NodeInUse(err error) response.Response {
	return response.ErrorResponse(StatusNodeInUse, err.Error())
}
