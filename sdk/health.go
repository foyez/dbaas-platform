package dbaas

import "context"

type HealthStatus struct {
	Status string `json:"status"`
}

// Health checks API availability
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	var status HealthStatus

	if err := c.do(ctx, "GET", "/health", nil, &status); err != nil {
		return nil, err
	}

	return &status, nil
}
