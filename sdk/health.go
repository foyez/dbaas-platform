package dbaas

import (
	"context"
	"net/http"
)

// HealthStatus is the API's health check response.
type HealthStatus struct {
	Status string `json:"status"`
}

// Health checks API availability. It needs no authentication.
func (c *Client) Health(ctx context.Context) (*HealthStatus, error) {
	var status HealthStatus

	if err := c.do(ctx, request{method: http.MethodGet, path: "/health", out: &status}); err != nil {
		return nil, err
	}

	return &status, nil
}
