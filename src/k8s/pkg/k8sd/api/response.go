package api

import (
	"github.com/canonical/lxd/lxd/response"
)

func InvalidNode(err error) response.Response {
	return response.ErrorResponse(517, err.Error())
}
