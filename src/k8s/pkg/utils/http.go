package utils

import (
	"encoding/json"
	"net/http"

	"github.com/canonical/lxd/lxd/response"
)

// JSONResponse marshals the response to JSON and sets the status code.
func JSONResponse(status int, v any) response.Response {
	b, _ := json.Marshal(v)
	return _response(status, b)
}

// Response writes the response and sets the status code.
func _response(status int, v []byte) response.Response {
	return response.ManualResponse(func(w http.ResponseWriter) error {
		w.WriteHeader(status)
		w.Write(v)
		w.Write([]byte("\n"))
		return nil
	})
}
