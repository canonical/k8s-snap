package utils

import (
	"encoding/json"
	"net/http"

	lxd "github.com/canonical/lxd/lxd/response"
)

// JSONResponse marshals the response to JSON and sets the status code.
func JSONResponse(status int, v any) lxd.Response {
	b, _ := json.Marshal(v)
	return response(status, b)
}

// Response writes the response and sets the status code.
func response(status int, v []byte) lxd.Response {
	return lxd.ManualResponse(func(w http.ResponseWriter) error {
		w.WriteHeader(status)
		w.Write(v)
		w.Write([]byte("\n"))
		return nil
	})
}
