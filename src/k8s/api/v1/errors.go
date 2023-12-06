package v1

import "fmt"

// GetError implements cluster.ResponseWithErrorMessage.
func (v *CreateTokenResponse) GetError() error {
	if v.Error != "" {
		return fmt.Errorf(v.Error)
	}
	return nil
}

// GetError implements cluster.ResponseWithErrorMessage.
func (v *CheckTokenResponse) GetError() error {
	if v.Error != "" {
		return fmt.Errorf(v.Error)
	}
	return nil
}
