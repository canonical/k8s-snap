package snapd

import (
	"encoding/json"
	"fmt"
	"io"
)

type snapdSnapInfoResponse struct {
	StatusCode int `json:"status-code"`
}

func (c *Client) GetSnapInfo(snap string) (*snapdSnapInfoResponse, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://localhost/v2/snaps/%s", snap))
	if err != nil {
		return nil, fmt.Errorf("failed to get snapd snap info: %w", err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("client: could not read response body: %w", err)
	}

	var snapInfoResponse snapdSnapInfoResponse
	if err := json.Unmarshal(resBody, &snapInfoResponse); err != nil {
		return nil, fmt.Errorf("client: could not unmarshal response body: %w", err)
	}

	return &snapInfoResponse, nil
}
