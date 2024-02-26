package errors

import (
	"errors"
	"fmt"
	"strings"

	v1 "github.com/canonical/k8s/api/v1"
)

var genericErrorMsgs = map[error]string{
	v1.ErrNotBootstrapped: "The cluster has not been initialized yet. Please call:\n\n sudo k8s bootstrap",
	v1.ErrConnectionFailed: "Unable to connect to the cluster. Verify k8s services are running:\n\n sudo snap services k8s\n\n" +
		"and see logs for more details:\n\n sudo journalctl -n 300 -u snap.k8s.k8sd\n\n",
	v1.ErrAPIServerFailed: "Unable to get Kubernetes API server endpoints. Verify k8s services are running:\n\n sudo snap services k8s\n\n" +
		"and see logs for more details:\n\n sudo journalctl -n 300 -u snap.k8s.kube-apiserver\n\n",
	v1.ErrTimeout: "Command timed out. See logs for more details:\n\n" +
		" sudo journalctl -n 300 -u snap.k8s.k8sd\n\n" +
		"You may increase the timeout with `--timeout 3m`.",
}

// Transform checks if the error returned by the server contains a known error message
// and transforms it into an user-friendly error message. The error type is lost when sending it over the wire,
// thus the error type cannot be checked.
// Transform is intended to be used with `defer`, therefore changing error in place.
func Transform(err *error, extraErrorMessages map[error]string) {
	if *err == nil {
		return
	}

	// Unknown error occured. Append the full error message to the result.
	if strings.Contains(strings.ToLower((*err).Error()), strings.ToLower(v1.ErrUnknown.Error())) {
		var prefix string
		if msg, ok := extraErrorMessages[v1.ErrUnknown]; ok {
			prefix = msg
		} else {
			prefix = genericErrorMsgs[v1.ErrUnknown]
		}
		*err = fmt.Errorf("%s%s", prefix, (*err).Error())
		return
	}

	for errorType, msg := range extraErrorMessages {
		if strings.Contains(strings.ToLower((*err).Error()), strings.ToLower(errorType.Error())) {
			*err = errors.New(msg)
			return
		}
	}

	for errorType, msg := range genericErrorMsgs {
		if strings.Contains(strings.ToLower((*err).Error()), strings.ToLower(errorType.Error())) {
			*err = errors.New(msg)
		}
	}
}
