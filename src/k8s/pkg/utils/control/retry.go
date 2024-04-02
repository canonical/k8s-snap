package control

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
