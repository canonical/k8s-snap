package csrsigning

import "time"

const (
	// requeueAfterSigningFailure is the time to requeue requests when any step of the signing process failed.
	requeueAfterSigningFailure = 3 * time.Second

	// requeueAfterWaitingForApproved is the amount of time to requeue requests if waiting for CSR to be approved.
	requeueAfterWaitingForApproved = 10 * time.Second

	// missingKeyFailedReason indicates the failure reason used when the CA
	// private key is not available to the controller.
	missingKeyFailedReason = "MissingCAKey"

	// missingKeyFailedMessage provides the failure message used when the
	// controller is unable to sign the CSR due to a missing CA private key.
	missingKeyFailedMessage = "The CSR could not be signed because the controller is missing the CA private key."
)
