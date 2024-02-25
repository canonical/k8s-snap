package v1

import "errors"

// ErrNotBootstrapped indicates that the cluster is not yet bootstrapped.
var ErrNotBootstrapped = errors.New("daemon not yet initialized")

// ErrAlreadyBootstrapped indicates that there is already a cluster bootstrapped on this node.
var ErrAlreadyBootstrapped = errors.New("cluster already bootstrapped")

// ErrInvalidJoinToken indicates that a node tried to join the cluster with an invalid token.
var ErrInvalidJoinToken = errors.New("failed to join cluster with the given join token")

// ErrTokenAlreadyCreated indicates that a token for this node was already created.
// TODO: Instead, return the already existing token.
var ErrTokenAlreadyCreated = errors.New("UNIQUE constraint failed: internal_token_records.name")

// ErrTimeout indicates that the action on the server took too long.
var ErrTimeout = errors.New("context deadline exceeded")

// ErrConnectionFailed indicates that a connection to the k8sd daemon could not be established.
var ErrConnectionFailed = errors.New("dial unix")

// ErrAPIServerFailed indicates that kube-apiserver endpoint(s) could not be determined.
var ErrAPIServerFailed = errors.New("failed to get kube-apiserver endpoints")

// ErrUnknown indicates that the server returns an unknown error.
var ErrUnknown = errors.New("unknown error")
