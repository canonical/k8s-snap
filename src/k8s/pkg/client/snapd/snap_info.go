package snapd

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type SnapInfoResult struct {
	InstallDate time.Time `json:"install-date"`
}

type SnapInfoResponse struct {
	StatusCode int            `json:"status-code"`
	Result     SnapInfoResult `json:"result"`
}

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

func (s SnapInfoResponse) HasInstallDate() bool {
	return !s.Result.InstallDate.IsZero()
}
