package cert

import (
	"crypto/rand"
	"math/big"
)

// generateSerialNumber returns a random number that can be used for the SerialNumber field in an x509 certificate.
func generateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}
