package p2pclient

import (
	"context"
	"encoding/json"
	"fmt"
)

// Plans returns a list of available plans.
func (c *Client) Plans(ctx context.Context) ([]Plan, error) {
	// public api, no need nonce and ts
	resp, err := c.request(ctx, PlansEndpoint.Method, PlansEndpoint.Uri, nil, "n1", 0)
	if err != nil {
		return nil, err
	}

	var plans []Plan
	err = json.Unmarshal(resp.Body, &plans)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Plans: %w", err)
	}

	return plans, nil
}
