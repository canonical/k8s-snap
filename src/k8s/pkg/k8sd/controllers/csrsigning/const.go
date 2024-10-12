package csrsigning

import "time"

const (
	// requeueAfterSigningFailure is the time to requeue requests when any step of the signing process failed.
	requeueAfterSigningFailure = 3 * time.Second

	// requeueAfterWaitingForApproved is the amount of time to requeue requests if waiting for CSR to be approved.
	requeueAfterWaitingForApproved = 10 * time.Second
)
