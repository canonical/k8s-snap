package utils

import (
	"encoding/json"
	"fmt"
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

// responseRenderer is a function that renders a response to the response writer.
type responseRenderer func(w http.ResponseWriter, r *http.Request) error

// manualResponseWithSignal creates a manual response that flushes the response to
// the client and signals completion on the given channel.
func manualResponseWithSignal(readyCh chan error, r *http.Request, renderer responseRenderer) lxd.Response {
	return lxd.ManualResponse(func(w http.ResponseWriter) (rerr error) {
		defer func() {
			readyCh <- rerr
			close(readyCh)
		}()

		if err := renderer(w, r); err != nil {
			return fmt.Errorf("failed to render response: %w", err)
		}

		f, ok := w.(http.Flusher)
		if !ok {
			return fmt.Errorf("ResponseWriter is not type http.Flusher")
		}

		f.Flush()
		return nil
	})
}

// SyncManualResponseWithSignal is a convenience wrapper for manualResponseWithSignal
// that renders a standard SyncResponse.
func SyncManualResponseWithSignal(req *http.Request, readyCh chan error, result any) lxd.Response {
	return manualResponseWithSignal(readyCh, req, func(w http.ResponseWriter, r *http.Request) error {
		return lxd.SyncResponse(true, result).Render(w, r)
	})
}
