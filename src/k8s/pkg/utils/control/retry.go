package control

// RetryFor will retry a given function for the given amount of times.
// RetryFor will not wait between retries. This is up to the retryFunc to handle.
func RetryFor(retryCount int, retryFunc func() error) error {
	var err error = nil
	for i := 0; i < retryCount; i++ {
		if err = retryFunc(); err != nil {
			continue
		}
		break
	}
	return err
}
