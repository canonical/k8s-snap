package utils

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

// IsLocalPortOpen checks if the given local port is already open or not.
func IsLocalPortOpen(port string) (bool, error) {
	// Without an address, Listen will listen on all addresses.
	if l, err := net.Listen("tcp", fmt.Sprintf(":%s", port)); errors.Is(err, syscall.EADDRINUSE) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		l.Close()
		return true, nil
	}
}
