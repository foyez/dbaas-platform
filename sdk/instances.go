package dbaas

import (
	"context"
	"fmt"
)

// InstancesService groups all instance-related API calls, accessed via
// client.Instances.<Method>(...).
type InstancesService struct {
	client *Client
}

type CreateInstanceInput struct {
	Name     string `json:"name"`
	Version  int    `json:"version"`
	Storage  string `json:"storage"`
	Username string `json:"username"`

	// IdempotencyKey is sent as a header, not the body - matches your
	// router's CORS config, which already allows this header.
	IdempotencyKey string `json:"-"`
}

func (s *InstancesService) Create(ctx context.Context, input CreateInstanceInput) (*Instance, error) {
	var out Instance

	headers := map[string]string{}
	if input.IdempotencyKey != "" {
		headers["Idempotency-Key"] = input.IdempotencyKey
	}

	if err := s.client.do(ctx, "POST", "/api/v1/instances", input, &out, headers); err != nil {
		return nil, fmt.Errorf("create instance: %w", err)
	}

	return &out, nil
}

func (s *InstancesService) Get(ctx context.Context, id string) (*Instance, error) {
	var out Instance

	if err := s.client.do(ctx, "GET", "/api/v1/instances/"+id, nil, &out, nil); err != nil {
		return nil, fmt.Errorf("get instance: %w", err)
	}

	return &out, nil
}

func (s *InstancesService) List(ctx context.Context) ([]*Instance, error) {
	var out []*Instance

	if err := s.client.do(ctx, "GET", "/api/v1/instances", nil, &out, nil); err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}

	return out, nil
}

type UpdateInstanceInput struct {
	Version *int    `json:"version,omitempty"`
	Storage *string `json:"storage,omitempty"`
}

func (s *InstancesService) Update(ctx context.Context, id string, input UpdateInstanceInput) (*Instance, error) {
	var out Instance

	if err := s.client.do(ctx, "PATCH", "/api/v1/instances/"+id, input, &out, nil); err != nil {
		return nil, fmt.Errorf("update instance: %w", err)
	}

	return &out, nil
}

func (s *InstancesService) Delete(ctx context.Context, id string) error {
	if err := s.client.do(ctx, "DELETE", "/api/v1/instances/"+id, nil, nil, nil); err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}

	return nil
}

func (s *InstancesService) Credentials(ctx context.Context, id string) (*InstanceCredentials, error) {
	var out InstanceCredentials

	if err := s.client.do(ctx, "GET", "/api/v1/instances/"+id+"/credentials", nil, &out, nil); err != nil {
		return nil, fmt.Errorf("get instance credentials: %w", err)
	}

	return &out, nil
}

func (s *InstancesService) Logs(ctx context.Context, id string) ([]LogLine, error) {
	var out struct {
		InstanceID string    `json:"instanceId"`
		Logs       []LogLine `json:"logs"`
	}

	if err := s.client.do(ctx, "GET", "/api/v1/instances/"+id+"/logs", nil, &out, nil); err != nil {
		return nil, fmt.Errorf("get instance logs: %w", err)
	}

	return out.Logs, nil
}
