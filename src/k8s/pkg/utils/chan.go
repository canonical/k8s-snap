package utils

// MaybeNotify pushes an empty struct to a channel, but does not block if that fails.
func MaybeNotify(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// MaybeReceive tries to receive from a channel, but does not block if that fails.
func MaybeReceive(ch <-chan struct{}) {
	select {
	case <-ch:
	default:
	}
}
