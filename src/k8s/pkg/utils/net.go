package utils

import (
	"errors"
	"net"
	"os"
	"syscall"
	"time"
)

// IsLocalPortOpen checks if the given local port is already open or not.
func IsLocalPortOpen(port string) (bool, error) {
	if err := checkPort("localhost", port, 500*time.Millisecond); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrDeadlineExceeded) || errors.Is(err, syscall.ECONNREFUSED) {
		return false, nil
	} else {
		// could not open due to error, couldn't check.
		return false, err
	}
}

func checkPort(host, port string, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
