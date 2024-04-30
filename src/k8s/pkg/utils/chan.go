package utils

// MaybeNotify pushes an empty struct to a channel, but does not block if that fails.
func MaybeNotify(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}
