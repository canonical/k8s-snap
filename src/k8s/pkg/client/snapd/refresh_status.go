package snapd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/canonical/k8s/pkg/k8sd/types"
)

type snapdChangeResponse struct {
	Result types.RefreshStatus `json:"result"`
}

func (c *Client) GetRefreshStatus(changeID string) (*types.RefreshStatus, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://localhost/v2/changes/%s", changeID))
	if err != nil {
		return nil, fmt.Errorf("failed to get snapd change status: %w", err)
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("client: could not read response body: %w", err)
	}

	var changeResponse snapdChangeResponse
	if err := json.Unmarshal(resBody, &changeResponse); err != nil {
		return nil, fmt.Errorf("client: could not unmarshal response body: %w", err)
	}

	return &changeResponse.Result, nil
}
