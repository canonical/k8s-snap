package metallb

import "time"

// SetWebhookEndpointsTimeoutForTest overrides webhookEndpointsTimeout and returns a restore function.
func SetWebhookEndpointsTimeoutForTest(d time.Duration) func() {
	prev := webhookEndpointsTimeout
	webhookEndpointsTimeout = d
	return func() {
		webhookEndpointsTimeout = prev
	}
}
