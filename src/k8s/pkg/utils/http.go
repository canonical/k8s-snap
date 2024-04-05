package utils

import (
	"encoding/json"
	"net/http"

	lxd "github.com/canonical/lxd/lxd/response"
)

// JSONResponse marshals the response to JSON and sets the status code.
func JSONResponse(status int, v any) lxd.Response {
	b, err := json.Marshal(v)
	if err != nil {
		return lxd.InternalError(err)
	}
	return response(status, b)
}

// Response writes the response and sets the status code.
func response(status int, v []byte) lxd.Response {
	return lxd.ManualResponse(func(w http.ResponseWriter) error {
		w.WriteHeader(status)
		_, err := w.Write(v)
		if err != nil {
			return err
		}
		w.Write([]byte("\n"))
		return nil
	})
}
