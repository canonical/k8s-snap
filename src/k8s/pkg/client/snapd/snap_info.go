package snapd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type SnapInfoResult struct {
	InstallDate time.Time `json:"install-date"`
	Revision    int       `json:"revision"`
}

type SnapInfoResponse struct {
	StatusCode int            `json:"status-code"`
	Result     SnapInfoResult `json:"result"`
}

// GetSnapInfo retrieves information about a snap from the snapd API.
func (c *Client) GetSnapInfo(snap string) (*SnapInfoResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://localhost/v2/snaps/%s", snap))
	if err != nil {
		return nil, fmt.Errorf("failed to get snapd snap info: %w", err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("client: could not read response body: %w", err)
	}

	var snapInfoResponse SnapInfoResponse
	if err := json.Unmarshal(resBody, &snapInfoResponse); err != nil {
		return nil, fmt.Errorf("client: could not unmarshal response body: %w", err)
	}

	return &snapInfoResponse, nil
}

// HasInstallDate checks if the InstallDate field is set in the SnapInfoResult.
// It returns true if the InstallDate is not zero, indicating that the snap is installed.
func (s SnapInfoResponse) HasInstallDate() bool {
	return !s.Result.InstallDate.IsZero()
}
